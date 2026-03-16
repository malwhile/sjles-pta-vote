package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
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

type CreatePollRequest struct {
	Question      string `json:"question"`
	DurationHours int    `json:"duration_hours,omitempty"`
}

func AdminNewPollHandler(resWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		// Parse JSON request body
		var req CreatePollRequest
		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			common.SendError(resWriter, "Invalid JSON request", http.StatusBadRequest)
			return
		}

		if req.Question == "" {
			common.SendError(resWriter, "Question is required", http.StatusBadRequest)
			return
		}

		// Use default duration if not provided
		durationHours := DEFAULT_POLL_DURATION_HOURS
		if req.DurationHours > 0 {
			durationHours = req.DurationHours
		}

		poll := models.Poll{
			Question:  req.Question,
			ExpiresAt: time.Now().Add(time.Duration(durationHours) * time.Hour).Format(common.DATE_FORMAT),
		}

		if _, err := CreatePoll(&poll); err != nil {
			common.SendError(resWriter, "Failed to create poll", http.StatusInternalServerError)
			return
		}

		common.SendSuccess(resWriter, map[string]bool{"success": true})

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

// GetAllPollsHandler returns all polls as a public endpoint (GET /api/polls)
func GetAllPollsHandler(resWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	polls, err := GetAllPolls()
	if err != nil {
		log.Printf("ERROR: Failed to get all polls: %v", err)
		common.SendError(resWriter, "Failed to get polls", http.StatusInternalServerError)
		return
	}

	common.SendSuccess(resWriter, polls)
}

// EditPollHandler updates an existing poll (PATCH /api/admin/polls/{id})
func EditPollHandler(resWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "PATCH" && request.Method != http.MethodPut {
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract poll ID from URL path
	parts := strings.Split(strings.TrimPrefix(request.URL.Path, "/api/admin/polls/"), "/")
	idStr := parts[0]
	if idStr == "" {
		common.SendError(resWriter, "Poll ID required", http.StatusBadRequest)
		return
	}

	pollID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.SendError(resWriter, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	// Parse JSON request body
	var req CreatePollRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		common.SendError(resWriter, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		common.SendError(resWriter, "Question is required", http.StatusBadRequest)
		return
	}

	db_conn, err := db.GetDB()
	if err != nil {
		log.Printf("ERROR: Failed to get database connection: %v", err)
		common.SendError(resWriter, "Database error", http.StatusInternalServerError)
		return
	}

	// Update poll question and/or expiration
	query := "UPDATE polls SET question = ?"
	args := []interface{}{req.Question}

	if req.DurationHours > 0 {
		query += ", expires_at = ?"
		args = append(args, time.Now().Add(time.Duration(req.DurationHours)*time.Hour).Format(common.DATE_FORMAT))
	}

	query += " WHERE id = ?"
	args = append(args, pollID)

	_, err = db_conn.Exec(query, args...)
	if err != nil {
		log.Printf("ERROR: Failed to update poll: %v", err)
		common.SendError(resWriter, "Failed to update poll", http.StatusInternalServerError)
		return
	}

	common.SendSuccess(resWriter, map[string]interface{}{"message": "Poll updated successfully"})
}

// DeletePollHandler deletes a poll (DELETE /api/admin/polls/{id})
func DeletePollHandler(resWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract poll ID from URL path
	parts := strings.Split(strings.TrimPrefix(request.URL.Path, "/api/admin/polls/"), "/")
	idStr := parts[0]
	if idStr == "" {
		common.SendError(resWriter, "Poll ID required", http.StatusBadRequest)
		return
	}

	pollID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.SendError(resWriter, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	db_conn, err := db.GetDB()
	if err != nil {
		log.Printf("ERROR: Failed to get database connection: %v", err)
		common.SendError(resWriter, "Database error", http.StatusInternalServerError)
		return
	}

	// Delete voters first (due to foreign key constraint)
	_, err = db_conn.Exec("DELETE FROM voters WHERE poll_id = ?", pollID)
	if err != nil {
		log.Printf("ERROR: Failed to delete voters: %v", err)
		common.SendError(resWriter, "Failed to delete poll", http.StatusInternalServerError)
		return
	}

	// Delete the poll
	result, err := db_conn.Exec("DELETE FROM polls WHERE id = ?", pollID)
	if err != nil {
		log.Printf("ERROR: Failed to delete poll: %v", err)
		common.SendError(resWriter, "Failed to delete poll", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("ERROR: Failed to get rows affected: %v", err)
		common.SendError(resWriter, "Failed to verify deletion", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		common.SendError(resWriter, "Poll not found", http.StatusNotFound)
		return
	}

	common.SendSuccess(resWriter, map[string]interface{}{"message": "Poll deleted successfully"})
}
