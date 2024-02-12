package phoenix

import (
	"bufio"
	"database/sql"
	"errors"
	"hash/crc32"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type phoenix struct {
	config *Config
	db     *sql.DB
	dbType dbType
}

func (p *phoenix) migrate() error {
	history, err := p.getHistory()
	if err != nil {
		if err := p.createHistoryTable(); err != nil {
			return err
		}
	}

	importRecords, err := p.getImportRecords()
	if err != nil {
		return err
	}

	err = p.executeMigration(history, importRecords)
	if err != nil {
		return err
	}

	return nil
}

func (p *phoenix) getHistory() ([]*historyEntry, error) {
	var result []*historyEntry
	rows, err := p.db.Query("SELECT * FROM " + p.config.TableName() + " ORDER BY version ASC;")
	if err != nil {

		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry historyEntry
		err = rows.Scan(&entry.installedRank, &entry.version, &entry.description, &entry.migrationType, &entry.script, &entry.checksum, &entry.installedBy, &entry.installedOn, &entry.executionTime, &entry.success)
		if err != nil {
			return []*historyEntry{}, err
		}
		result = append(result, &entry)
	}
	return result, nil
}

func (p *phoenix) createHistoryTable() error {
	var err error
	switch p.dbType {
	case Postgres:
		fallthrough
	case Mysql:
		_, err = p.db.Exec("CREATE TABLE " + p.config.TableName() + " (installed_rank INT, version VARCHAR(50), description VARCHAR(200), type VARCHAR(20), script VARCHAR(1000), checksum NUMERIC, installed_by VARCHAR(100), installed_on TIMESTAMP, execution_time NUMERIC, success BOOLEAN);")
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *phoenix) getImportRecords() ([]*importRecord, error) {
	var collectedRecords []*importRecord
	if err := filepath.WalkDir(p.config.ImportFolder, p.collectImports(&collectedRecords)); err != nil {
		return nil, err
	}
	sort.Slice(collectedRecords, func(p, j int) bool {
		return collectedRecords[p].getVersion() < collectedRecords[j].getVersion()
	})
	return collectedRecords, nil
}

func (p *phoenix) collectImports(collector *[]*importRecord) func(path string, d fs.DirEntry, err error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return errors.New("Import path does not exist")
		}
		if d.IsDir() {
			return nil
		}
		re := regexp.MustCompile(`V(\d+)__.+\.sql`)
		if !re.MatchString(d.Name()) {
			log.Println("Skipping file", d.Name())
			return nil
		}
		version := re.FindStringSubmatch(d.Name())[1]
		if len(version) == 0 {
			return errors.New("File does not version naming convention")
		}

		fileContent, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		*collector = append(*collector, &importRecord{
			sqlCommands: removeComments(strings.SplitAfter(string(fileContent), ";")),
			name:        d.Name(),
			checksum:    p.checksum(fileContent),
		})
		return nil
	}
}

func (p *phoenix) checksum(input []byte) uint32 {
	table := crc32.MakeTable(crc32.IEEE)
	return crc32.Checksum(input, table)
}

func (p *phoenix) executeMigration(history []*historyEntry, records []*importRecord) error {
	maxVersion := -1

	var err error
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	for index, record := range records {
		if index < len(history) {
			maxVersion, err = p.validateHistoryEntry(history[index], record)
			if err != nil {
				return err
			}
		} else {
			currentVersion, err := strconv.Atoi(record.getVersion())
			if err != nil {
				return err
			}
			if currentVersion <= maxVersion {
				return errors.New("version conflict")
			}
			err = p.importRecord(tx, record)
			if err != nil {
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					return rollbackErr
				}
				return err
			}
			maxVersion = currentVersion
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
func (p *phoenix) importRecord(tx *sql.Tx, record *importRecord) error {
	startTime := time.Now()
	for _, command := range record.sqlCommands {
		_, err := tx.Exec(command)
		if err != nil {
			return err
		}
	}
	duration := time.Since(startTime)
	rank, err := record.getRank()
	if err != nil {
		return err
	}
	currentUser, err := p.getCurrentUser()
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO "+p.config.TableName()+" (installed_rank, version, description, type, script, checksum, installed_by, installed_on, execution_time, success) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);",
		rank,
		record.getVersion(),
		record.getDescription(),
		"SQL",
		record.name,
		record.checksum,
		currentUser,
		time.Now().UTC(),
		duration.Milliseconds(),
		true)
	if err != nil {
		return err
	}
	return nil
}

func (p *phoenix) getCurrentUser() (string, error) {
	currentUser := new(string)
	var err error
	switch p.dbType {
	case Postgres:
		err = p.db.QueryRow("SELECT current_user;").Scan(currentUser)
	case Mysql:
		err = p.db.QueryRow("SELECT CURRENT_USER();").Scan(currentUser)
	}
	if err != nil {
		return "", err

	}
	return *currentUser, nil
}

func (p *phoenix) validateHistoryEntry(entry *historyEntry, record *importRecord) (int, error) {
	if err := p.validate(entry, record); err != nil {
		return 0, err
	}
	maxVersion, err := strconv.Atoi(entry.version)
	if err != nil {
		log.Fatal(err)
	}
	return maxVersion, nil
}

func (p *phoenix) validate(historyEntry *historyEntry, record *importRecord) error {
	if historyEntry.version != record.getVersion() {
		return errors.New("Version mismatch")
	}
	if historyEntry.checksum != record.checksum {
		return errors.New("Checksum mismatch")
	}
	return nil
}

func removeComments(commands []string) []string {
	result := make([]string, 0)
	re := regexp.MustCompile(`(#.*)`)
	for _, command := range commands {
		cleanedCommand := ""
		scanner := bufio.NewScanner(strings.NewReader(command))
		for scanner.Scan() {
			line := scanner.Text()
			cleanedCommand += strings.TrimLeft(re.ReplaceAllLiteralString(line, ""), "")
			if len(cleanedCommand) > 0 {
				cleanedCommand += "\n"
			}
		}
		if len(cleanedCommand) != 0 {
			result = append(result, cleanedCommand[:len(cleanedCommand)-1])
		}
	}
	return result
}
