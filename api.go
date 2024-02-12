package phoenix

import "database/sql"

type dbType string

const (
	Postgres dbType = "postgres"
	Mysql    dbType = "mysql"
)

type ConfigFunc func(config *Config)

func WithImportFolder(folder string) ConfigFunc {
	return func(config *Config) {
		config.ImportFolder = folder
	}
}

func WithTableName(table string) ConfigFunc {
	return func(config *Config) {
		config.Table = table
	}
}

func WithSchema(schema string) ConfigFunc {
	return func(config *Config) {
		config.SchemaName = schema
	}
}

func defaultConfig() Config {
	return Config{
		ImportFolder: "phoenix",
		Table:        "phoenix_history",
		SchemaName:   "",
	}
}

type Config struct {
	ImportFolder string
	Table        string
	SchemaName   string
}

func (c *Config) TableName() string {
	schema := c.SchemaName
	if schema != "" {
		schema += "."
	}
	return schema + c.Table
}

func Rise(db *sql.DB, dbType dbType, configFns ...ConfigFunc) {
	config := defaultConfig()
	for _, fn := range configFns {
		fn(&config)
	}
	p := phoenix{
		config: &config,
		db:     db,
		dbType: dbType,
	}
	if err := p.migrate(); err != nil {
		panic(err)
	}
}
