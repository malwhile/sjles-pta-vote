package db

import (
	"os"
	"testing"

	"go-sjles-pta-vote/server/config"
)

func TestConnect(t *testing.T) {
	tmp_db, err := os.CreateTemp("", "vote_test.*.db")
	if err != nil {
		t.Errorf(`Failed to create temporary db file: %v`, err)
	}

	init_conf := &config.Config{
		DBPath: string(tmp_db.Name()),
	}
	config.SetConfig(init_conf)

	defer os.Remove(tmp_db.Name())
	tmp_db.Close()

	if _, err := Connect(); err != nil {
		t.Errorf(`Failed to create the database: %v`, err)
	}

	defer Close()
}