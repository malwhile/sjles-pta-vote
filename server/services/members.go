package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/db"
	"go-sjles-pta-vote/server/logging"
)

type Member struct {
	Name  string
	Email string
}

const BATCH_SIZE = 100
const CVS_FILE_FIELD = "members.csv"

func AdminMembersHandler(resWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		resWriter.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var year int
	var err error

	if err = request.ParseMultipartForm(10 << 20); err != nil {
		common.SendError(resWriter, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	year_from_form := request.FormValue("year")
	if year_from_form == "" {
		common.SendError(resWriter, "Year is required", http.StatusBadRequest)
		return
	} else {
		year, err = strconv.Atoi(year_from_form)
		if err != nil {
			common.SendError(resWriter, "Invalid year", http.StatusBadRequest)
			return
		}
	}

	file, _, err := request.FormFile(CVS_FILE_FIELD)
	if err != nil {
		common.SendError(resWriter, "Failed to read " + CVS_FILE_FIELD + " file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		common.SendError(resWriter, "Failed to read " + CVS_FILE_FIELD + " file", http.StatusInternalServerError)
		return
	}

	memberCount, err := ParseMembersFromBytes(year, fileBytes)
	if err != nil {
		logging.Errorf("failed to parse members from CSV: %v", err)
		common.SendError(resWriter, "Failed to parse members from CSV", http.StatusBadRequest)
		return
	}

	// Extract admin user from token if available
	authHeader := request.Header.Get("Authorization")
	adminUser := "unknown"
	if authHeader != "" && len(authHeader) > 7 {
		if username, err := VerifyAuthToken(authHeader[7:]); err == nil {
			adminUser = username
		}
	}

	logging.Audit("UPLOAD_MEMBERS", adminUser, fmt.Sprintf("year=%d count=%d", year, memberCount), true)

	resWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(resWriter).Encode(map[string]interface{}{common.SUCCESS: true, "count": memberCount})
}

func AdminMembersView(resWriter http.ResponseWriter, request *http.Request) {
	yearStr := request.URL.Query().Get("year")
	if yearStr == "" {
		common.SendError(resWriter, "Year is required", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		common.SendError(resWriter, "Invalid year", http.StatusBadRequest)
		return
	}

	members, err := GetMembersByYear(year)
	if err != nil {
		logging.Errorf("failed to get members for year %d: %v", year, err)
		common.SendError(resWriter, "Failed to get members", http.StatusInternalServerError)
		return
	}

	// Return empty array if no members found
	if members == nil {
		members = []Member{}
	}

	common.SendSuccess(resWriter, members)
}

func ParseMembersFromBytes(year int, fileBytes []byte) (int, error) {
	reader := csv.NewReader(strings.NewReader(string(fileBytes)))
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record
	records, err := reader.ReadAll()
	if err != nil {
		return 0, errors.Wrap(err, "failed to read CSV from bytes")
	}

	if len(records) == 0 {
		return 0, errors.New("CSV file is empty")
	}

	var members []Member
	skippedCount := 0

	for i, record := range records {
		if i == 0 {
			continue // Skip the first line (column headers)
		}

		// Validate minimum required fields
		if len(record) < 4 {
			skippedCount++
			logging.Warnf("Row %d: skipped due to insufficient fields (required 4, got %d)", i+1, len(record))
			continue
		}

		firstName := strings.TrimSpace(record[1])
		lastName := strings.TrimSpace(record[2])
		email := strings.TrimSpace(record[3])

		// Validate required fields are not empty
		if firstName == "" && lastName == "" {
			skippedCount++
			logging.Warnf("Row %d: skipped due to empty name fields", i+1)
			continue
		}

		if email == "" {
			skippedCount++
			logging.Warnf("Row %d: skipped due to empty email field", i+1)
			continue
		}

		members = append(members, Member{
			Name:  fmt.Sprintf("%s %s", firstName, lastName),
			Email: email,
		})

		if len(record) < 30 {
			continue
		}

		email2 := strings.TrimSpace(record[27])
		if email2 != "" {
			firstName2 := strings.TrimSpace(record[29])
			lastName2 := strings.TrimSpace(record[28])

			members = append(members, Member{
				Name:  fmt.Sprintf("%s %s", firstName2, lastName2),
				Email: email2,
			})
		}
	}

	err = saveMember(year, members)
	if err != nil {
		return 0, err
	}
	return len(members), nil
}

func saveMember(year int, members []Member) error {
	insertMembersQuery := `
		INSERT OR REPLACE INTO members (email, member_name, school_year)
		VALUES ($1, $2, $3)
	`
	logging.Infof("starting to save %d members for year %d", len(members), year)

	db_conn, err := db.Connect()
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	tx, err := db_conn.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	stmt, err := tx.Prepare(insertMembersQuery)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "failed to prepare statement")
	}
	defer stmt.Close()

	for index, member := range members {
		_, err = stmt.Exec(member.Email, member.Name, year)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "failed to execute insert")
		}

		if (index+1) % BATCH_SIZE == 0 {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "failed to commit transaction")
			}

			tx, err = db_conn.Begin()
			if err != nil {
				return errors.Wrap(err, "failed to begin new transaction")
			}

			stmt, err = tx.Prepare(insertMembersQuery)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "failed to prepare new statement")
			}
		}
	}

	return tx.Commit()
}

func GetMembersByYear(year int) ([]Member, error) {
	query := `
		SELECT member_name, email
		FROM members
		WHERE school_year = $1
		ORDER BY member_name ASC
	`

	db_conn, err := db.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	rows, err := db_conn.Query(query, year)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var members []Member

	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.Name, &member.Email); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "row iteration error")
	}

	return members, nil
}