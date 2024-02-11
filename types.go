package phoenix

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

type historyEntry struct {
	installedRank int
	version       string
	description   string
	migrationType string
	script        string
	checksum      uint32
	installedBy   string
	installedOn   time.Time
	executionTime int
	success       bool
}

type importRecord struct {
	sqlCommands []string
	name        string
	checksum    uint32
}

func (r *importRecord) getVersion() string {
	re := regexp.MustCompile(`V(\d+)__.+`)
	return re.FindStringSubmatch(r.name)[1]
}

func (r *importRecord) getDescription() string {
	re := regexp.MustCompile(`V\d+__(.+)`)
	return re.FindStringSubmatch(r.name)[1]
}

func (r *importRecord) getRank() (uint32, error) {
	rank, err := strconv.Atoi(r.getVersion())
	if err != nil {
		return 0, errors.New("Could not identify rank")
	}
	return uint32(rank), nil
}
