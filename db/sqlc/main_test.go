package db

import (
	"context"
	"log"
	"os"
	"tkbank/util"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	configDB, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Fatal("Cannot parse config: ", err)
	}

	configDB.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	testDB, err = pgxpool.NewWithConfig(context.Background(), configDB)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
