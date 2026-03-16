package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/db"
	"go-sjles-pta-vote/server/models"
)

var (
	ErrQuestionAlreadyExists = errors.New("Question already exists")
	ErrQuestionDoesntExist   = errors.New("Question does not exist yet")
	ErrVoterAlreadyVoted     = errors.New("Voter already voted")
	ErrPollNotFound          = errors.New("Poll not found")
	ErrFailedToUpdateVote    = errors.New("Failed to update vote")
	ErrFailedToDeletePoll    = errors.New("Failed to delete poll")
)

const (
	DEFAULT_POLL_DURATION_HOURS = 24
)

type PollRequest struct {
	PollId int64 `json:"poll_id"`
}

func AdminNewPollHandler(resWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		question := request.FormValue("question")
		if question == "" {
			common.SendError(resWriter, "Question is required", http.StatusBadRequest)
			return
		}

		durationHours := DEFAULT_POLL_DURATION_HOURS
		if durationStr := request.FormValue("duration"); durationStr != "" {
			var err error
			durationHours, err = strconv.Atoi(durationStr)
			if err != nil {
				common.SendError(resWriter, "Invalid duration", http.StatusBadRequest)
				return
			}
		}

		poll := models.Poll{
			Question:  question,
			ExpiresAt: time.Now().Add(time.Duration(durationHours) * time.Hour).Format(common.DATE_FORMAT),
		}

		if _, err := CreatePoll(&poll); err != nil {
			common.SendError(resWriter, "Failed to create poll", http.StatusInternalServerError)
			return
		}

		resWriter.WriteHeader(http.StatusOK)
		json.NewEncoder(resWriter).Encode(map[string]bool{common.SUCCESS: true})

	default:
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func CreatePoll(poll *models.Poll) (int64, error) {
	db_conn, err := db.Connect()
	if err != nil {
		log.Printf("Failed to connect to database: %s", err.Error())
		return -1, err
	}
	defer db.Close()

	get_stmt, err := db_conn.Prepare(`
		SELECT id 
			FROM polls 
			WHERE question == $1
	`)
	if err != nil {
		log.Printf("%s", err.Error())
		return -1, err
	}
	defer get_stmt.Close()

	var id int
	err = get_stmt.QueryRow(poll.Question).Scan(&id)
	if err == nil {
		// Question already exists
		return -1, ErrQuestionAlreadyExists
	} else if err != sql.ErrNoRows {
		// Some other database error
		log.Printf("%s", err.Error())
		return -1, err
	}
	// Question doesn't exist, proceed to insert

	stmt, err := db_conn.Prepare(`
		INSERT INTO polls (
			question,
			expires_at
		) VALUES (
		 	$1,
			$2
		) RETURNING ID;
	`)

	if err != nil {
		log.Printf("%s", err.Error())
		return -1, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(poll.Question, poll.ExpiresAt)
	if err != nil {
		log.Printf("%s", err.Error())
		return -1, err
	}

	new_poll_id, err := res.LastInsertId()
	return new_poll_id, err
}

func AdminViewPollHandler(resWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		var req PollRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			common.SendError(resWriter, "Invalid JSON", http.StatusBadRequest)
			return
		}

		poll, err := GetPollById(req.PollId)
		if err == ErrPollNotFound {
			common.SendError(resWriter, "Poll not found", http.StatusNotFound)
			return
		} else if err != nil {
			common.SendError(resWriter, "Failed to get poll", http.StatusInternalServerError)
			return
		}

		resWriter.WriteHeader(http.StatusOK)
		json.NewEncoder(resWriter).Encode(poll)
	case http.MethodGet:
		var polls []models.Poll
		var err error
		question := request.FormValue("question")
		if question == "" {
			polls, err = GetAllPolls()
			if err != nil {
				common.SendError(resWriter, "Failed to get polls", http.StatusInternalServerError)
				return
			}
		} else {
			poll, err := GetPollByQuestion(question)
			if err != nil {
				common.SendError(resWriter, "Failed to get poll question "+question, http.StatusInternalServerError)
				return
			}
			polls = append(polls, *poll)
		}

		err = json.NewEncoder(resWriter).Encode(polls)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
			common.SendError(resWriter, "Failed to encode polls", http.StatusInternalServerError)
			return
		}
	default:
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetAllPolls() ([]models.Poll, error) {
	// Use optimized utility function that fetches polls with voters in single query
	pollPtrs, err := QueryAllPolls()
	if err != nil {
		return nil, err
	}

	// Convert pointers to values for backward compatibility
	polls := make([]models.Poll, 0, len(pollPtrs))
	for _, p := range pollPtrs {
		polls = append(polls, *p)
	}

	return polls, nil
}

func GetPollByQuestion(question string) (*models.Poll, error) {
	// Use optimized utility function that fetches poll with voters in single query
	return QueryPollByQuestion(question)
}

func GetPollById(id int64) (*models.Poll, error) {
	// Use optimized utility function that fetches poll with voters in single query
	return QueryPollByID(id)
}

func GetAndCreatePollByQuestion(question string) (*models.Poll, error) {
	new_poll, err := GetPollByQuestion(question)

	if err == ErrPollNotFound {
		create_poll := &models.Poll{
			Question:  question,
			ExpiresAt: time.Now().Add(time.Hour * 10).Format(common.DATE_FORMAT),
		}

		if _, err = CreatePoll(create_poll); err != nil {
			return nil, err
		}

		return GetPollByQuestion(question)
	} else if err != nil {
		log.Printf("%s", err.Error())
		return nil, err
	} else {
		return new_poll, err
	}
}

func SetVote(vote *models.Vote) error {
	db_conn, err := db.Connect()
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer db.Close()

	set_voter_stmt, err := db_conn.Prepare(`
		INSERT OR IGNORE INTO voters
			(poll_id, voter_email)
		VALUES ($1, $2)
	`)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer set_voter_stmt.Close()

	res, err := set_voter_stmt.Exec(vote.PollId, vote.Email)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	} else {
		rows_changed, err := res.RowsAffected()
		if rows_changed != 1 {
			return ErrVoterAlreadyVoted
		} else if err != nil {
			log.Printf("%s", err.Error())
			return err
		}
	}

	is_voter_member_stmt, err := db_conn.Prepare(`
		SELECT 1
		FROM members
		WHERE email == $1
	`)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer is_voter_member_stmt.Close()

	var member_check int64
	is_member := true
	err = is_voter_member_stmt.QueryRow(vote.Email).Scan(&member_check)
	if err == sql.ErrNoRows {
		is_member = false
	} else if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	// Member column name is not dependant on user input
	//	So it's ok to put it directly in the query
	member_column_name := "member_"
	if !is_member {
		member_column_name = "non_" + member_column_name
	}

	if vote.Vote {
		member_column_name += "yes_votes"
	} else {
		member_column_name += "no_votes"
	}

	add_vote_stmt, err := db_conn.Prepare(`
		UPDATE polls
		SET ` + member_column_name + ` = ` + member_column_name + ` + 1
		WHERE id == $1
	`)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer add_vote_stmt.Close()

	res, err = add_vote_stmt.Exec(vote.PollId)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	if num, err := res.RowsAffected(); num != 1 {
		return ErrFailedToUpdateVote
	} else if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	return nil
}

// Delete a poll by name
func DeletePollByQuestion(question string) error {
	db_conn, err := db.Connect()
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer db.Close()

	delete_votes_stmt, err := db_conn.Prepare(`
		DELETE FROM voters
		WHERE poll_id IN (
			SELECT id 
			FROM polls 
			WHERE question == $1
		)
	`)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer delete_votes_stmt.Close()

	_, err = delete_votes_stmt.Exec(question)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	delete_poll_stmt, err := db_conn.Prepare(`
		DELETE FROM polls
		WHERE question == $1
	`)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer delete_poll_stmt.Close()

	res, err := delete_poll_stmt.Exec(question)
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	if num, err := res.RowsAffected(); num != 1 {
		return ErrFailedToDeletePoll
	} else if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	return nil
}

func CreatePollIgnore(poll *models.Poll) error {
	db_conn, err := db.Connect()
	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}
	defer db.Close()

	stmt, err := db_conn.Prepare(`
		INSERT OR IGNORE INTO polls (
			question,
			expires_at,
			member_yes_votes,
			member_no_votes,
			non_member_yes_votes,
			non_member_no_votes,
			created_at,
			updated_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
	`)

	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(
		poll.Question,
		poll.ExpiresAt,
		poll.MemberYes,
		poll.MemberNo,
		poll.NonMemberYes,
		poll.NonMemberNo,
		poll.CreatedAt,
		poll.UpdatedAt,
	)

	if err != nil {
		log.Printf("%s", err.Error())
		return err
	}

	return nil
}
