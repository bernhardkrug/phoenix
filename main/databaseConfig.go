package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func connectToDB() (*gorm.DB, error) {
	postgresConfig := postgres.Config{
		DSN:                  `host=127.0.0.1 user=incrementus password=incrementus dbname=incrementus port=5432 sslmode=disable`,
		PreferSimpleProtocol: true,
	}
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ddhub.", // schema name
			SingularTable: false,
		},
	}
	db, err := gorm.Open(postgres.New(postgresConfig), &gormConfig)
	if err != nil {
		return nil, err
	}
	return db, nil
}
