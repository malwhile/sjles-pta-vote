package db

import (
	"database/sql"
	"log"
	"sync"

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

// Global database connection pool - thread-safe singleton
var (
	dbInstance *sql.DB
	dbMutex    sync.Mutex
	initialized bool
)

// GetDB returns the shared database connection pool
func GetDB() (*sql.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbInstance != nil {
		return dbInstance, nil
	}

	// Initialize connection pool
	db_config := config.GetConfig()

	db, err := sql.Open("sqlite", db_config.DBPath)
	if err != nil {
		log.Printf("ERROR: Failed to open database: %v", err)
		return nil, err
	}

	// Initialize schema
	_, err = db.Exec(build_db_query)
	if err != nil {
		log.Printf("ERROR: Failed to initialize database schema: %v", err)
		_ = db.Close()
		return nil, err
	}

	dbInstance = db
	initialized = true
	log.Printf("INFO: Database initialized successfully")

	return dbInstance, nil
}

// Connect is deprecated - use GetDB instead
// Kept for backward compatibility
func Connect() (*sql.DB, error) {
	return GetDB()
}

// Close closes the shared database connection
func Close() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbInstance != nil {
		err := dbInstance.Close()
		if err != nil {
			log.Printf("ERROR: Failed to close database: %v", err)
		} else {
			log.Printf("INFO: Database connection closed")
		}
		dbInstance = nil
		initialized = false
	}
}

// ClearDatabase clears all data from the database (for testing)
func ClearDatabase() error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM voters; DELETE FROM polls; DELETE FROM members;")
	if err != nil {
		log.Printf("ERROR: Failed to clear database: %v", err)
		return err
	}

	log.Printf("INFO: Database cleared")
	return nil
}

// ResetDB resets the database singleton (for testing)
func ResetDB() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if dbInstance != nil {
		_ = dbInstance.Close()
		dbInstance = nil
		initialized = false
	}
}
