package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/db"
	"go-sjles-pta-vote/server/models"
	"go-sjles-pta-vote/server/services"
)

var setupDB bool

func voteHandler(resWriter http.ResponseWriter, request *http.Request) {
	var vote models.Vote
	if err := json.NewDecoder(request.Body).Decode(&vote); err != nil {
		common.SendError(resWriter, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := services.SetVote(&vote); err != nil {
		if err == services.ErrVoterAlreadyVoted {
			common.SendError(resWriter, "Already voted", http.StatusConflict)
			return
		}
		common.SendError(resWriter, "Failed to set vote", http.StatusInternalServerError)
		return
	}

	resWriter.WriteHeader(http.StatusOK)
}

func apiPollsMethodHandler(resWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "PATCH", "PUT":
		services.EditPollHandler(resWriter, request)
	case http.MethodDelete:
		services.DeletePollHandler(resWriter, request)
	default:
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func pollsIDHandler(resWriter http.ResponseWriter, request *http.Request) {
	parts := strings.Split(strings.TrimPrefix(request.URL.Path, "/api/polls/"), "/")
	idStr := parts[0]
	log.Printf("Received request for poll ID: %s", idStr)
	if idStr == "" {
		common.SendError(resWriter, "Poll ID not provided", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	log.Printf("Fetching poll with ID: %d", id)
	if err != nil {
		common.SendError(resWriter, "Invalid poll ID", http.StatusBadRequest)
		return
	}

	poll, err := services.GetPollById(id)
	log.Printf("Retrieved poll: %+v, error: %v", poll, err)
	if err == services.ErrPollNotFound {
		common.SendError(resWriter, "Poll not found", http.StatusNotFound)
		return
	} else if err != nil {
		common.SendError(resWriter, "Failed to get poll", http.StatusInternalServerError)
		return
	}

	log.Printf("Sending poll data: %+v", poll)
	resWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(resWriter).Encode(poll)
}

func adminLoginHandler(resWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		var loginReq services.LoginRequest
		if err := json.NewDecoder(request.Body).Decode(&loginReq); err != nil {
			common.SendError(resWriter, "Invalid JSON", http.StatusBadRequest)
			return
		}

		isValid, err := services.ValidateAdminLogin(loginReq.Username, loginReq.Password)
		if err != nil || !isValid {
			common.SendError(resWriter, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		token, err := services.GenerateAuthToken(loginReq.Username)
		if err != nil {
			common.SendError(resWriter, "Failed to generate auth token", http.StatusInternalServerError)
			return
		}

		resWriter.WriteHeader(http.StatusOK)
		json.NewEncoder(resWriter).Encode(services.LoginResponse{
			Success: true,
			Token:   token,
		})
	default:
		common.SendError(resWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func initDatabase() error {
	// Clear existing polls first
	if err := db.ClearDatabase(); err != nil {
		return fmt.Errorf("failed to clear polls: %v", err)
	}

	polls := []models.Poll{
		{
			ID:           1,
			Question:     "Should we increase the budget?",
			MemberYes:    rand.Int63n(50),
			MemberNo:     rand.Int63n(50),
			NonMemberYes: rand.Int63n(20),
			NonMemberNo:  rand.Int63n(20),
			TotalVotes:   int(rand.Int63n(100)),
			WhoVoted:     []string{"email1@example.com", "email2@example.com", "email3@example.com", "email4@example.com"},
			CreatedAt:    time.Now().Format(time.RFC3339),
			UpdatedAt:    time.Now().Format(time.RFC3339),
			ExpiresAt:    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:           2,
			Question:     "Should we hire more staff?",
			MemberYes:    rand.Int63n(50),
			MemberNo:     rand.Int63n(50),
			NonMemberYes: rand.Int63n(20),
			NonMemberNo:  rand.Int63n(20),
			TotalVotes:   int(rand.Int63n(100)),
			WhoVoted:     []string{"email1@example.com", "email2@example.com", "email3@example.com", "email4@example.com"},
			CreatedAt:    time.Now().Format(time.RFC3339),
			UpdatedAt:    time.Now().Format(time.RFC3339),
			ExpiresAt:    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
		{
			ID:           3,
			Question:     "Should we renovate the building?",
			MemberYes:    rand.Int63n(50),
			MemberNo:     rand.Int63n(50),
			NonMemberYes: rand.Int63n(20),
			NonMemberNo:  rand.Int63n(20),
			TotalVotes:   int(rand.Int63n(100)),
			WhoVoted:     []string{"email1@example.com", "email2@example.com", "email3@example.com", "email4@example.com"},
			CreatedAt:    time.Now().Format(time.RFC3339),
			UpdatedAt:    time.Now().Format(time.RFC3339),
			ExpiresAt:    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		},
	}

	for _, poll := range polls {
		if err := services.CreatePollIgnore(&poll); err != nil {
			return fmt.Errorf("failed to create poll %d: %v", poll.ID, err)
		}
	}
	return nil
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check if setupdb flag is present
	flag.BoolVar(&setupDB, "setupdb", false, "Initialize database with sample data")
	flag.Parse()

	if setupDB {
		if err := initDatabase(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
	}

	http.HandleFunc("/api/vote", voteHandler)
	http.HandleFunc("/api/polls", services.GetAllPollsHandler)           // GET - list all polls
	http.HandleFunc("/api/polls/", pollsIDHandler)                        // GET - get poll by ID
	http.HandleFunc("/api/admin/new-poll", services.AdminNewPollHandler)
	http.HandleFunc("/api/admin/view-polls", services.AdminViewPollHandler)
	http.HandleFunc("/api/admin/login", adminLoginHandler)
	http.HandleFunc("/api/admin/logout", services.LogoutHandler)          // POST - logout
	http.HandleFunc("/api/admin/polls/", apiPollsMethodHandler)           // PATCH/DELETE - edit/delete polls
	http.HandleFunc("/api/admin/members", services.AdminMembersHandler)
	http.HandleFunc("/api/admin/members/view", services.AdminMembersView)

	buildPath := filepath.Join(".", "client", "build")
	fs := http.FileServer(http.Dir(buildPath))

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(filepath.Join(buildPath, r.URL.Path)); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(buildPath, "index.html"))
	}))

	log.Printf("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
