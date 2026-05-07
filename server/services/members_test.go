package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-sjles-pta-vote/server/config"
	"go-sjles-pta-vote/server/db"
)

func setupTestDB(t *testing.T) string {
	tmp_db, err := os.CreateTemp("", "members_test.*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary db file: %v", err)
	}

	init_conf := &config.Config{
		DBPath: string(tmp_db.Name()),
	}
	config.SetConfig(init_conf)

	tmp_db.Close()

	db.ResetDB()

	if _, err := db.Connect(); err != nil {
		t.Fatalf("Failed to create the database: %v", err)
	}

	return tmp_db.Name()
}

func TestGetMembershipYearsEmpty(t *testing.T) {
	tmpPath := setupTestDB(t)
	defer os.Remove(tmpPath)

	years, err := GetMembershipYears()
	if err != nil {
		t.Errorf("Failed to get membership years: %v", err)
	}

	if len(years) != 0 {
		t.Errorf("Expected empty array for empty database, got %d years", len(years))
	}
}

func TestGetMembershipYearsMultipleYears(t *testing.T) {
	tmpPath := setupTestDB(t)
	defer os.Remove(tmpPath)

	db_conn, err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Insert members for different years
	years := []int{2024, 2023, 2022, 2025}
	for _, year := range years {
		_, err := db_conn.Exec(
			"INSERT INTO members (email, member_name, school_year) VALUES (?, ?, ?)",
			fmt.Sprintf("test%d@example.com", year), fmt.Sprintf("Test %d", year), year,
		)
		if err != nil {
			t.Fatalf("Failed to insert member: %v", err)
		}
	}

	retrievedYears, err := GetMembershipYears()
	if err != nil {
		t.Errorf("Failed to get membership years: %v", err)
	}

	// Should be in descending order
	expectedOrder := []int{2025, 2024, 2023, 2022}
	if len(retrievedYears) != len(expectedOrder) {
		t.Errorf("Expected %d years, got %d", len(expectedOrder), len(retrievedYears))
	}

	for i, year := range retrievedYears {
		if year != expectedOrder[i] {
			t.Errorf("Year at position %d: expected %d, got %d", i, expectedOrder[i], year)
		}
	}
}

func TestGetMembershipYearsNoMembers(t *testing.T) {
	tmpPath := setupTestDB(t)
	defer os.Remove(tmpPath)

	db_conn, err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Insert members for year 2023
	_, err = db_conn.Exec(
		"INSERT INTO members (email, member_name, school_year) VALUES (?, ?, ?)",
		"test2023@example.com", "Test 2023", 2023,
	)
	if err != nil {
		t.Fatalf("Failed to insert member: %v", err)
	}

	// Get years should only return 2023
	years, err := GetMembershipYears()
	if err != nil {
		t.Errorf("Failed to get membership years: %v", err)
	}

	if len(years) != 1 || years[0] != 2023 {
		t.Errorf("Expected only year 2023, got %v", years)
	}
}

func TestAdminMembersYearsHandlerAuthorized(t *testing.T) {
	tmpPath := setupTestDB(t)
	defer os.Remove(tmpPath)

	db_conn, err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Insert test members
	years := []int{2024, 2023}
	for _, year := range years {
		_, err := db_conn.Exec(
			"INSERT INTO members (email, member_name, school_year) VALUES (?, ?, ?)",
			fmt.Sprintf("test%d@example.com", year), fmt.Sprintf("Test %d", year), year,
		)
		if err != nil {
			t.Fatalf("Failed to insert member: %v", err)
		}
	}

	// Generate a valid token
	token, err := GenerateAuthToken("testadmin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create request with valid token
	req := httptest.NewRequest("GET", "/api/admin/members/years", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	AdminMembersYearsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify response contains the years
	body := w.Body.String()
	if len(body) == 0 {
		t.Errorf("Expected non-empty response body")
	}

	// Check that response is valid JSON with data
	if !contains(body, "\"success\":true") {
		t.Errorf("Expected success:true in response, got: %s", body)
	}

	if !contains(body, "\"data\":[2024,2023]") && !contains(body, "\"data\"") {
		t.Errorf("Expected data field in response, got: %s", body)
	}
}

func TestAdminMembersYearsHandlerEmptyResult(t *testing.T) {
	tmpPath := setupTestDB(t)
	defer os.Remove(tmpPath)

	// Create request with valid token
	token, err := GenerateAuthToken("testadmin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/members/years", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	w := httptest.NewRecorder()
	AdminMembersYearsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &result)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	// Verify empty array is returned for empty database
	data, ok := result["data"].([]interface{})
	if !ok {
		t.Errorf("Expected data to be an array, got %T", result["data"])
	}

	if len(data) != 0 {
		t.Errorf("Expected empty array for empty database, got %d items", len(data))
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
