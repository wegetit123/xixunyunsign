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

	// Create the school_info table (updated to include city_name and city_id)
	createSchoolTableSQL := `
    CREATE TABLE IF NOT EXISTS school_info (
        school_id TEXT PRIMARY KEY,
        school_name TEXT,
        city_name TEXT,  -- Added city_name
        city_id TEXT     -- Added city_id
    );
    `
	_, err = db.Exec(createSchoolTableSQL)
	if err != nil {
		return err
	}

	// Create schedules table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schedules (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        account TEXT,
        address TEXT,
        latitude TEXT,
        longitude TEXT,
        province TEXT,
        city TEXT,
        remark TEXT,
        comment TEXT,
        cron_expr TEXT,
        enabled INTEGER DEFAULT 1
    )`)
	if err != nil {
		return fmt.Errorf("创建 schedules 表失败: %v", err)
	}

	return nil
}

// SaveSchoolInfo saves or updates the school information in the database.
func SaveSchoolInfo(cityName, cityID, schoolID, schoolName string) error {
	if db == nil {
		if err := InitDB(); err != nil {
			return err
		}
	}

	insertSQL := `
    INSERT INTO school_info (city_name, city_id, school_id, school_name)
    VALUES (?, ?, ?, ?)
    ON CONFLICT(school_id) DO UPDATE SET
        school_name = excluded.school_name,
        city_name = excluded.city_name,
        city_id = excluded.city_id;
    `

	_, err := db.Exec(insertSQL, cityName, cityID, schoolID, schoolName)
	return err
}

// FetchAndSaveSchoolData fetches the school data from the given API and saves it to the database.
func FetchAndSaveSchoolData() error {
	// Make the GET request to the API
	resp, err := http.Get("https://api.xixunyun.com/login/schoolmap")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			CityName string       `json:"name"`
			CityId   string       `json:"id"`
			Schools  []SchoolInfo `json:"list"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Loop through the data and save each school info
	for _, group := range result.Data {
		for _, school := range group.Schools {
			if err := SaveSchoolInfo(group.CityName, group.CityId, school.SchoolID, school.SchoolName); err != nil {
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

	// Use SQL LIKE for fuzzy matching
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
		return false, fmt.Errorf("查询 school_info 表时发生错误: %v", err)
	}
	return count == 0, nil
}
