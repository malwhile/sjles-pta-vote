package services

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"go-sjles-pta-vote/server/config"
	"go-sjles-pta-vote/server/db"
	"go-sjles-pta-vote/server/models"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-=`~!@#$%^&*()_+[]\\;',./{}|:\"<>?"

var new_members = []struct {
	email       string
	member_name string
}{
	{"test1@mail.me", "test1"},
	{"test2@mail.me", "test2"},
	{"test3@mail.me", "test3"},
	{"test4@mail.me", "test4"},
	{"test5@mail.me", "test5"},
	{"test6@mail.me", "test6"},
	{"test7@mail.me", "test7"},
	{"test100@mail.me", "test100"},
	{"test101@mail.me", "test101"},
	{"test102@mail.me", "test102"},
	{"test103@mail.me", "test103"},
	{"test104@mail.me", "test104"},
	{"test105@mail.me", "test105"},
}

var new_polls = []struct {
	question             string
	member_yes_votes     int64
	member_no_votes      int64
	non_member_yes_votes int64
	non_member_no_votes  int64
}{
	{"ques1", 1, 2, 3, 4},
	{"ques2", 3, 2, 4, 5},
	{"ques3", 4, 3, 6, 5},
}

var new_voters = []struct {
	poll_id     int64
	voter_email string
}{
	{1, "test1@mail.me"},
	{1, "test2@mail.me"},
	{1, "test3@mail.me"},
	{1, "test10@mail.me"},
	{1, "test11@mail.me"},
	{1, "test12@mail.me"},
	{1, "test13@mail.me"},
	{1, "test14@mail.me"},
	{1, "test15@mail.me"},
	{1, "test16@mail.me"},
	{2, "test1@mail.me"},
	{2, "test2@mail.me"},
	{2, "test3@mail.me"},
	{2, "test4@mail.me"},
	{2, "test5@mail.me"},
	{2, "test10@mail.me"},
	{2, "test11@mail.me"},
	{2, "test12@mail.me"},
	{2, "test13@mail.me"},
	{2, "test14@mail.me"},
	{2, "test15@mail.me"},
	{2, "test16@mail.me"},
	{2, "test17@mail.me"},
	{2, "test18@mail.me"},
	{3, "test1@mail.me"},
	{3, "test2@mail.me"},
	{3, "test3@mail.me"},
	{3, "test4@mail.me"},
	{3, "test5@mail.me"},
	{3, "test6@mail.me"},
	{3, "test7@mail.me"},
	{3, "test10@mail.me"},
	{3, "test11@mail.me"},
	{3, "test12@mail.me"},
	{3, "test13@mail.me"},
	{3, "test14@mail.me"},
	{3, "test15@mail.me"},
	{3, "test16@mail.me"},
	{3, "test17@mail.me"},
	{3, "test18@mail.me"},
	{3, "test19@mail.me"},
	{3, "test20@mail.me"},
}

func RandString(length int) string {
	rand_bytes := make([]byte, length)
	for rand_index := range length {
		rand_bytes[rand_index] = charset[rand.Intn(len(charset))]
	}
	return string(rand_bytes)
}

func PreLoadDB() error {
	db_conn, err := db.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	// Insert members
	for i := range new_members {
		_, err := db_conn.Exec(`INSERT INTO members (email, member_name, school_year) VALUES (?, ?, ?)`, new_members[i].email, new_members[i].member_name, 2023)
		if err != nil {
			return err
		}
	}

	// Insert polls
	for i := range new_polls {
		result, err := db_conn.Exec(`INSERT INTO polls (question, member_yes_votes, member_no_votes, non_member_yes_votes, non_member_no_votes, expires_at) VALUES (?, ?, ?, ?, ?, ?)`, new_polls[i].question, new_polls[i].member_yes_votes, new_polls[i].member_no_votes, new_polls[i].non_member_yes_votes, new_polls[i].non_member_no_votes, time.Now().Add(time.Hour*10).Format("2006-01-02 15:04:05"))
		if err != nil {
			return err
		}
		_, err = result.LastInsertId()
		if err != nil {
			return err
		}
	}

	// Insert voters
	for i := range new_voters {
		_, err := db_conn.Exec(`INSERT INTO voters (poll_id, voter_email) VALUES (?, ?)`, new_voters[i].poll_id, new_voters[i].voter_email)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestCreatePoll(t *testing.T) {
	parameters := []struct {
		question    string
		table_index int64
	}{
		{RandString(10) + "1", 1},
		{RandString(10) + "2", 2},
		{RandString(10) + "3", 3},
		{"\"" + RandString(10) + "4", 4},
		{"\\\"" + RandString(10) + "5", 5},
		{"'" + RandString(10) + "6", 6},
		{";" + RandString(10) + "7", 7},
		{"\\" + RandString(10) + "8", 8},
	}

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

	// Reset database singleton to use new test database
	db.ResetDB()

	if _, err := db.Connect(); err != nil {
		t.Errorf(`Failed to create the database: %v`, err)
	}

	for i := range parameters {
		create_poll := &models.Poll{
			Question:  parameters[i].question,
			ExpiresAt: time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05"),
		}

		new_poll_id, err := CreatePoll(create_poll)

		if err != nil {
			t.Errorf(`Failed to create new poll %s: %v`, parameters[i].question, err)
		}

		if new_poll_id == -1 {
			t.Errorf(`Failed to insert %s into table`, parameters[i].question)
		}

		if new_poll_id != parameters[i].table_index {
			t.Errorf(`Incorrect increment in index for %s: expected %d != %d`, parameters[i].question, parameters[i].table_index, new_poll_id)
		}
	}
}

func TestAlreadyExists(t *testing.T) {
	question := "TestQuestion"

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

	// Reset database singleton to use new test database
	db.ResetDB()

	if _, err := db.Connect(); err != nil {
		t.Errorf(`Failed to create the database: %v`, err)
	}

	create_poll := &models.Poll{
		Question:  question,
		ExpiresAt: time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05"),
	}

	new_poll, err := CreatePoll(create_poll)

	if err != nil {
		t.Errorf(`Failed to create new poll %s: %v`, question, err)
	}

	if new_poll == -1 {
		t.Errorf(`Failed to insert %s into table`, question)
	}

	new_poll, err = CreatePoll(create_poll)

	if err != ErrQuestionAlreadyExists {
		t.Errorf(`Should have failed adding %s as it already exists`, question)
	}
}

func TestGetPollByQuestion(t *testing.T) {
	question := "TestQuestion"

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

	// Reset database singleton to use new test database
	db.ResetDB()

	if _, err := db.Connect(); err != nil {
		t.Errorf(`Failed to create the database: %v`, err)
	}

	create_poll := &models.Poll{
		Question:  question,
		ExpiresAt: time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05"),
	}

	new_poll, err := CreatePoll(create_poll)

	if err != nil {
		t.Errorf(`Failed to create new poll %s: %v`, question, err)
	}

	if new_poll == -1 {
		t.Errorf(`Failed to insert %s into table`, question)
	}

	get_poll, err := GetPollByQuestion(question)

	if err != nil {
		t.Errorf(`Failed to get the poll %s: %v`, question, err)
	}

	if get_poll.Question != question {
		t.Errorf(`Questions don't match: expected %s: recieved %s`, question, get_poll.Question)
	}
}

func TestGetCreatePollByQuestion(t *testing.T) {
	parameters := []struct {
		question    string
		table_index int64
	}{
		{RandString(10) + "1", 1},
		{RandString(10) + "2", 2},
		{RandString(10) + "3", 3},
		{"\"" + RandString(10) + "4", 4},
		{"'" + RandString(10) + "5", 5},
		{";" + RandString(10) + "6", 6},
	}

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

	// Reset database singleton to use new test database
	db.ResetDB()

	if _, err := db.Connect(); err != nil {
		t.Errorf(`Failed to create the database: %v`, err)
	}

	for i := range parameters {
		new_poll, err := GetAndCreatePollByQuestion(parameters[i].question)

		if err != nil {
			t.Errorf(`Failed to create new poll %s: %v`, parameters[i].question, err)
		}

		if new_poll == nil {
			t.Errorf(`Failed to insert %s into table`, parameters[i].question)
		}

		if new_poll.ID != parameters[i].table_index {
			t.Errorf(`Incorrect increment in index for %s: expected %d != %d`, parameters[i].question, parameters[i].table_index, new_poll.ID)
		}

		if new_poll.Question != parameters[i].question {
			t.Errorf(`Incorrect question returned: Expected %s != %s`, parameters[i].question, new_poll.Question)
		}
	}
}

func TestSetVote(t *testing.T) {
	// Preload the database with members, polls, and voters
	tmp_db, err := os.CreateTemp("", "vote_test.*.db")
	if err != nil {
		t.Errorf("Failed to create temporary database: %v", err)
	}
	defer os.Remove(tmp_db.Name())

	init_conf := &config.Config{
		DBPath: string(tmp_db.Name()),
	}
	config.SetConfig(init_conf)

	tmp_db.Close()

	// Reset database singleton to use new test database
	db.ResetDB()

	err = PreLoadDB()
	if err != nil {
		t.Errorf("Failed to preload database: %v", err)
	}

	// Add a non-member vote
	random_email := RandString(10) + "@mail.me"
	vote := &models.Vote{
		PollId: 1,
		Email:  random_email,
		Vote:   true,
	}
	err = SetVote(vote)
	if err != nil {
		t.Errorf("Failed to set non-member vote: %v", err)
	}

	// Add a member vote
	member_email := "test100@mail.me"
	vote = &models.Vote{
		PollId: 1,
		Email:  member_email,
		Vote:   true,
	}
	err = SetVote(vote)
	if err != nil {
		t.Errorf("Failed to set member vote: %v", err)
	}

	// Verify the votes were added correctly
	voters, err := models.GetVoters(1) // Use GetVoters from models
	if err != nil {
		t.Errorf("Failed to get voters: %v", err)
	}

	expected_non_member_votes := 4 + 1 // Original non-member votes + new non-member vote
	expected_member_votes := 3 + 1     // Original member votes + new member vote

	for _, voter := range voters {
		if voter.Email == random_email && voter.YesVote {
			expected_non_member_votes--
		} else if voter.Email == member_email && voter.YesVote {
			expected_member_votes--
		}
	}

	if expected_non_member_votes != 5 || expected_member_votes != 4 {
		t.Errorf("Expected %d non-member votes and %d member votes, but got %d non-member votes and %d member votes", 4+1, 3+1, expected_non_member_votes, expected_member_votes)
	}
}

func TestVoterAlreadyVoted(t *testing.T) {
	// Preload the database with members, polls, and voters
	tmp_db, err := os.CreateTemp("", "vote_test.*.db")
	if err != nil {
		t.Errorf("Failed to create temporary database: %v", err)
	}
	defer os.Remove(tmp_db.Name())

	init_conf := &config.Config{
		DBPath: string(tmp_db.Name()),
	}
	config.SetConfig(init_conf)

	err = PreLoadDB()
	if err != nil {
		t.Errorf("Failed to preload database: %v", err)
	}

	// Add a non-member vote
	random_email := RandString(10) + "@mail.me"
	vote := &models.Vote{
		PollId: 1,
		Email:  random_email,
		Vote:   true,
	}
	err = SetVote(vote)
	if err != nil {
		t.Errorf("Failed to set non-member vote: %v", err)
	}

	// Add a member vote
	member_email := "test100@mail.me"
	vote = &models.Vote{
		PollId: 1,
		Email:  member_email,
		Vote:   true,
	}
	err = SetVote(vote)
	if err != nil {
		t.Errorf("Failed to set member vote: %v", err)
	}

	// Attempt to add another non-member vote
	vote = &models.Vote{
		PollId: 1,
		Email:  random_email,
		Vote:   true,
	}
	err = SetVote(vote)
	if err != ErrVoterAlreadyVoted {
		t.Errorf("Expected ErrVoterAlreadyVoted, but got %v", err)
	}

	// Attempt to add another member vote
	vote = &models.Vote{
		PollId: 1,
		Email:  member_email,
		Vote:   true,
	}
	err = SetVote(vote)
	if err != ErrVoterAlreadyVoted {
		t.Errorf("Expected ErrVoterAlreadyVoted, but got %v", err)
	}
}

func TestDeletePollByQuestion(t *testing.T) {
	// Preload the database with members, polls, and voters
	tmp_db, err := os.CreateTemp("", "vote_test.*.db")
	if err != nil {
		t.Errorf("Failed to create temporary database: %v", err)
	}
	defer os.Remove(tmp_db.Name())

	init_conf := &config.Config{
		DBPath: string(tmp_db.Name()),
	}
	config.SetConfig(init_conf)

	err = PreLoadDB()
	if err != nil {
		t.Errorf("Failed to preload database: %v", err)
	}

	// Get a question from the new_polls array
	testQuestion := new_polls[0].question

	// Delete the poll by question
	err = DeletePollByQuestion(testQuestion)
	if err != nil {
		t.Errorf("Failed to delete poll by question: %v", err)
	}

	// Verify that the poll was deleted
	_, err = GetPollByQuestion(testQuestion)
	if err == nil {
		t.Errorf("Expected error when getting deleted poll, but got none")
	} else if err != ErrPollNotFound {
		t.Errorf("Expected ErrPollNotFound, but got %v", err)
	}
}
