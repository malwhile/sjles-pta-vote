package services

import (
	"database/sql"
	"log"
	"strings"

	"go-sjles-pta-vote/server/db"
	"go-sjles-pta-vote/server/models"
)

// QueryPollByID fetches a single poll by ID using JOIN to avoid N+1 queries
func QueryPollByID(pollID int64) (*models.Poll, error) {
	database, err := db.GetDB()
	if err != nil {
		log.Printf("ERROR: Failed to get database connection: %v", err)
		return nil, err
	}

	// Use LEFT JOIN to fetch poll and voters in single query
	query := `
		SELECT
			p.id,
			p.question,
			p.member_yes_votes,
			p.member_no_votes,
			p.non_member_yes_votes,
			p.non_member_no_votes,
			p.created_at,
			p.updated_at,
			p.expires_at,
			GROUP_CONCAT(v.voter_email) as who_voted
		FROM polls p
		LEFT JOIN voters v ON p.id = v.poll_id
		WHERE p.id = ?
		GROUP BY p.id
	`

	stmt, err := database.Prepare(query)
	if err != nil {
		log.Printf("ERROR: Failed to prepare statement: %v", err)
		return nil, err
	}
	defer stmt.Close()

	poll := &models.Poll{}
	var whoVotedStr sql.NullString

	err = stmt.QueryRow(pollID).Scan(
		&poll.ID,
		&poll.Question,
		&poll.MemberYes,
		&poll.MemberNo,
		&poll.NonMemberYes,
		&poll.NonMemberNo,
		&poll.CreatedAt,
		&poll.UpdatedAt,
		&poll.ExpiresAt,
		&whoVotedStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPollNotFound
		}
		log.Printf("ERROR: Failed to query poll by ID: %v", err)
		return nil, err
	}

	// Parse voters from GROUP_CONCAT result
	if whoVotedStr.Valid && whoVotedStr.String != "" {
		poll.WhoVoted = strings.Split(whoVotedStr.String, ",")
	} else {
		poll.WhoVoted = []string{}
	}

	return poll, nil
}

// QueryPollByQuestion fetches a single poll by question using JOIN to avoid N+1 queries
func QueryPollByQuestion(question string) (*models.Poll, error) {
	database, err := db.GetDB()
	if err != nil {
		log.Printf("ERROR: Failed to get database connection: %v", err)
		return nil, err
	}

	query := `
		SELECT
			p.id,
			p.question,
			p.member_yes_votes,
			p.member_no_votes,
			p.non_member_yes_votes,
			p.non_member_no_votes,
			p.created_at,
			p.updated_at,
			p.expires_at,
			GROUP_CONCAT(v.voter_email) as who_voted
		FROM polls p
		LEFT JOIN voters v ON p.id = v.poll_id
		WHERE p.question = ?
		GROUP BY p.id
	`

	stmt, err := database.Prepare(query)
	if err != nil {
		log.Printf("ERROR: Failed to prepare statement: %v", err)
		return nil, err
	}
	defer stmt.Close()

	poll := &models.Poll{}
	var whoVotedStr sql.NullString

	err = stmt.QueryRow(question).Scan(
		&poll.ID,
		&poll.Question,
		&poll.MemberYes,
		&poll.MemberNo,
		&poll.NonMemberYes,
		&poll.NonMemberNo,
		&poll.CreatedAt,
		&poll.UpdatedAt,
		&poll.ExpiresAt,
		&whoVotedStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPollNotFound
		}
		log.Printf("ERROR: Failed to query poll by question: %v", err)
		return nil, err
	}

	// Parse voters from GROUP_CONCAT result
	if whoVotedStr.Valid && whoVotedStr.String != "" {
		poll.WhoVoted = strings.Split(whoVotedStr.String, ",")
	} else {
		poll.WhoVoted = []string{}
	}

	return poll, nil
}

// QueryAllPolls fetches all polls using JOIN to avoid N+1 queries
func QueryAllPolls() ([]*models.Poll, error) {
	database, err := db.GetDB()
	if err != nil {
		log.Printf("ERROR: Failed to get database connection: %v", err)
		return nil, err
	}

	query := `
		SELECT
			p.id,
			p.question,
			p.member_yes_votes,
			p.member_no_votes,
			p.non_member_yes_votes,
			p.non_member_no_votes,
			p.created_at,
			p.updated_at,
			p.expires_at,
			GROUP_CONCAT(v.voter_email) as who_voted
		FROM polls p
		LEFT JOIN voters v ON p.id = v.poll_id
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`

	rows, err := database.Query(query)
	if err != nil {
		log.Printf("ERROR: Failed to query all polls: %v", err)
		return nil, err
	}
	defer rows.Close()

	polls := []*models.Poll{}

	for rows.Next() {
		poll := &models.Poll{}
		var whoVotedStr sql.NullString

		err = rows.Scan(
			&poll.ID,
			&poll.Question,
			&poll.MemberYes,
			&poll.MemberNo,
			&poll.NonMemberYes,
			&poll.NonMemberNo,
			&poll.CreatedAt,
			&poll.UpdatedAt,
			&poll.ExpiresAt,
			&whoVotedStr,
		)

		if err != nil {
			log.Printf("ERROR: Failed to scan poll row: %v", err)
			continue
		}

		// Parse voters from GROUP_CONCAT result
		if whoVotedStr.Valid && whoVotedStr.String != "" {
			poll.WhoVoted = strings.Split(whoVotedStr.String, ",")
		} else {
			poll.WhoVoted = []string{}
		}

		polls = append(polls, poll)
	}

	if err = rows.Err(); err != nil {
		log.Printf("ERROR: Error iterating poll rows: %v", err)
		return nil, err
	}

	return polls, nil
}

// CheckVoterExists checks if a voter has already voted on a poll
func CheckVoterExists(pollID int64, email string) (bool, error) {
	database, err := db.GetDB()
	if err != nil {
		log.Printf("ERROR: Failed to get database connection: %v", err)
		return false, err
	}

	query := "SELECT 1 FROM voters WHERE poll_id = ? AND voter_email = ? LIMIT 1"
	stmt, err := database.Prepare(query)
	if err != nil {
		log.Printf("ERROR: Failed to prepare statement: %v", err)
		return false, err
	}
	defer stmt.Close()

	var exists int
	err = stmt.QueryRow(pollID, email).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("ERROR: Failed to check voter exists: %v", err)
		return false, err
	}

	return true, nil
}
