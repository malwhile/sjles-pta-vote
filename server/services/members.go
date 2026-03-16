package services

import (
	"encoding/csv"
	"fmt"
	"log"
	"strings"
	"net/http"
	"strconv"
	"io/ioutil"
	"encoding/json"

	"github.com/pkg/errors"

	"go-sjles-pta-vote/server/common"
	"go-sjles-pta-vote/server/db"
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

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		common.SendError(resWriter, "Failed to read " + CVS_FILE_FIELD + " file", http.StatusInternalServerError)
		return
	}

	if err = ParseMembersFromBytes(year, fileBytes); err != nil {
		common.SendError(resWriter, "Failed to parse members from CSV", http.StatusBadRequest)
		return
	}

	resWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(resWriter).Encode(map[string]bool{common.SUCCESS: true})
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
		common.SendError(resWriter, "Failed to get members", http.StatusInternalServerError)
		return
	}

	resWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(resWriter).Encode(map[string]interface{}{
		common.SUCCESS: true,
		"members": members,
	})
}

func ParseMembersFromBytes(year int, fileBytes []byte) error {
	reader := csv.NewReader(strings.NewReader(string(fileBytes)))
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record
	records, err := reader.ReadAll()
	if err != nil {
		return errors.Wrap(err, "failed to read CSV from bytes")
	}

	var members []Member

	for i, record := range records {
		if i == 0 {
			continue // Skip the first line (column headers)
		}
		if len(record) < 4 {
			continue
		}

		firstName := strings.TrimSpace(record[1])
		lastName := strings.TrimSpace(record[2])
		email := strings.TrimSpace(record[3])

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

	return saveMember(year, members)
}

func saveMember(year int, members []Member) error {
	insertMembersQuery := `
		INSERT OR REPLACE INTO members (email, member_name, school_year)
		VALUES ($1, $2, $3)
	`
	log.Printf("Starting to save %d members for year %d", len(members), year)

	db_conn, err := db.Connect()
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer db_conn.Close()

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
	defer db_conn.Close()

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