package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

var db *sql.DB

// SchoolInfo represents the structure of school data.
type SchoolInfo struct {
	SchoolID   string `json:"school_id"`
	SchoolName string `json:"school_name"`
}

// InitDB initializes the database and creates the required tables.
func InitDB() error {
	// Set the database file path to "config.db"
	dbPath := "config.db"

	// Open the database connection
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create the users table
	createUserTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        account TEXT PRIMARY KEY,
        password TEXT,
        token TEXT,
        latitude TEXT,
        longitude TEXT,
        bind_phone TEXT,
        user_number TEXT,
        user_name TEXT,
        school_id INT,
        sex TEXT,
        class_name TEXT,
        entrance_year TEXT,
        graduation_year TEXT
    );
    `
	_, err = db.Exec(createUserTableSQL)
	if err != nil {
		return err
	}

	// Create the school_info table
	createSchoolTableSQL := `
    CREATE TABLE IF NOT EXISTS school_info (
        school_id TEXT PRIMARY KEY,
        school_name TEXT
    );
    `
	_, err = db.Exec(createSchoolTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// SaveSchoolInfo saves or updates the school information in the database.
func SaveSchoolInfo(schoolID, schoolName string) error {
	if db == nil {
		if err := InitDB(); err != nil {
			return err
		}
	}

	insertSQL := `
    INSERT INTO school_info (school_id, school_name)
    VALUES (?, ?)
    ON CONFLICT(school_id) DO UPDATE SET
        school_name = excluded.school_name;
    `
	_, err := db.Exec(insertSQL, schoolID, schoolName)
	return err
}

// FetchAndSaveSchoolData fetches the school data from the given API and saves it to the database.
func FetchAndSaveSchoolData() error {
	// Make the GET request to the API
	resp, err := http.Get("https://oss-resume.xixunyun.com/school_map/app202412.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			Groud   string       `json:"groud"`
			Schools []SchoolInfo `json:"schools"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Loop through the data and save each school info
	for _, group := range result.Data {
		for _, school := range group.Schools {
			if err := SaveSchoolInfo(school.SchoolID, school.SchoolName); err != nil {
				log.Printf("Error saving school %s: %v", school.SchoolName, err)
			}
		}
	}

	return nil
}

// SaveUser saves or updates a user in the database.
func SaveUser(account, password, token, latitude, longitude, bindPhone, userNumber, userName string, schoolID float64, sex, className, entranceYear, graduationYear string) error {
	if db == nil {
		if err := InitDB(); err != nil {
			return err
		}
	}
	insertSQL := `
    INSERT INTO users (
        account, password, token, latitude, longitude, bind_phone, 
        user_number, user_name, school_id, sex, class_name, entrance_year, graduation_year
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(account) DO UPDATE SET 
        password = excluded.password,
        token = excluded.token,
        latitude = excluded.latitude,
        longitude = excluded.longitude,
        bind_phone = excluded.bind_phone,
        user_number = excluded.user_number,
        user_name = excluded.user_name,
        school_id = excluded.school_id,
        sex = excluded.sex,
        class_name = excluded.class_name,
        entrance_year = excluded.entrance_year,
        graduation_year = excluded.graduation_year;
    `
	_, err := db.Exec(insertSQL, account, password, token, latitude, longitude, bindPhone, userNumber, userName, schoolID, sex, className, entranceYear, graduationYear)
	return err
}

// GetUser retrieves user information from the database by account.
func GetUser(account string) (token, latitude, longitude string, err error) {
	if db == nil {
		if err := InitDB(); err != nil {
			return "", "", "", err
		}
	}
	querySQL := `SELECT token, latitude, longitude FROM users WHERE account = ?;`
	row := db.QueryRow(querySQL, account)
	err = row.Scan(&token, &latitude, &longitude)
	return
}

// CloseDB closes the database connection.
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// UpdateCoordinates updates the latitude and longitude for a given account.
func UpdateCoordinates(account, latitude, longitude string) error {
	if db == nil {
		if err := InitDB(); err != nil {
			return err
		}
	}
	updateSQL := `
        UPDATE users
        SET latitude = ?, longitude = ?
        WHERE account = ?;
    `
	_, err := db.Exec(updateSQL, latitude, longitude, account)
	return err
}

// GetAdditionalUserData retrieves additional user data for constructing the query parameters.
func GetAdditionalUserData(account string) (map[string]string, error) {
	if db == nil {
		if err := InitDB(); err != nil {
			return nil, err
		}
	}
	querySQL := `
        SELECT entrance_year, graduation_year, school_id
        FROM users
        WHERE account = ?;
    `
	row := db.QueryRow(querySQL, account)

	var entranceYear, graduateYear, schoolID string
	err := row.Scan(&entranceYear, &graduateYear, &schoolID)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"entrance_year":   entranceYear,
		"graduation_year": graduateYear,
		"school_id":       schoolID,
	}, nil
}

// SearchSchoolID searches for all school IDs by school name using fuzzy matching.
func SearchSchoolID(schoolName string) ([]SchoolInfo, error) {
	if db == nil {
		if err := InitDB(); err != nil {
			return nil, err
		}
	}

	// 使用 SQL LIKE 进行模糊查询，%符号表示匹配任意字符
	querySQL := `SELECT school_id, school_name FROM school_info WHERE school_name LIKE ?;`
	likeName := "%" + schoolName + "%"

	rows, err := db.Query(querySQL, likeName)
	if err != nil {
		return nil, fmt.Errorf("查询学校ID时发生错误: %v", err)
	}
	defer rows.Close()

	var schools []SchoolInfo
	for rows.Next() {
		var school SchoolInfo
		if err := rows.Scan(&school.SchoolID, &school.SchoolName); err != nil {
			return nil, fmt.Errorf("读取查询结果时发生错误: %v", err)
		}
		schools = append(schools, school)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历查询结果时发生错误: %v", err)
	}

	return schools, nil
}

func IsSchoolInfoTableEmpty() (bool, error) {
	var count int
	querySQL := `SELECT COUNT(*) FROM school_info;`
	err := db.QueryRow(querySQL).Scan(&count)
	if err != nil {
		log.Printf("Error checking school_info table: %v", err)
		return false, err
	}
	return count == 0, nil
}
