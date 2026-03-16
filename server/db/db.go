package db

import (
	"database/sql"
	"log"

	"go-sjles-pta-vote/server/config"

	_ "github.com/glebarez/go-sqlite"
)

var build_db_query string = `
CREATE TABLE IF NOT EXISTS polls (
    id INTEGER PRIMARY KEY,
    question TEXT NOT NULL,
    member_yes_votes UNSIGNED INT NOT NULL DEFAULT 0,
    member_no_votes UNSIGNED INT NOT NULL DEFAULT 0,
    non_member_yes_votes UNSIGNED INT NOT NULL DEFAULT 0,
    non_member_no_votes UNSIGNED INT NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME
);

CREATE TABLE IF NOT EXISTS voters (
    poll_id UNSIGNED INT NOT NULL,
    voter_email TEXT NOT NULL,
    FOREIGN KEY (poll_id) REFERENCES polls(id),
    PRIMARY KEY (poll_id, voter_email)
);

CREATE TABLE IF NOT EXISTS members (
    email TEXT NOT NULL,
    member_name TEXT,
    school_year UNSIGNED INT NOT NULL,
    PRIMARY KEY (email, school_year)
);
`

var db *sql.DB

func Connect() (*sql.DB, error) {
	db_config := config.GetConfig()

	db, err := sql.Open("sqlite", db_config.DBPath)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, err
	}

	_, err = db.Exec(build_db_query)
	if err != nil {
		log.Printf("Error updating schema: %v", err)
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func Close() {
	if db != nil {
		err := db.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

func ClearDatabase() error {
	mydb, err := Connect()
	if err != nil {
		return err
	}

	_, err = mydb.Exec("DELETE FROM voters; DELETE FROM polls; DELETE FROM members;")
	if err != nil {
		log.Printf("Error clearing database: %v", err)
		return err
	}

	return nil
}
