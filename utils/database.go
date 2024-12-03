package utils

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the database and creates the users table if it doesn't exist.
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
	createTableSQL := `
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
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// SaveUser saves or updates a user in the database.
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
