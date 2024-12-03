package utils

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() error {
	// 将数据库文件路径设置为当前目录下的 "config.db"
	dbPath := "config.db"

	// 打开数据库连接
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// 创建表
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        account TEXT PRIMARY KEY,
        token TEXT,
        latitude TEXT,
        longitude TEXT
    );
    `
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func SaveUser(account, token, latitude, longitude string) error {
	if db == nil {
		if err := InitDB(); err != nil {
			return err
		}
	}
	insertSQL := `
    INSERT INTO users (account, token, latitude, longitude)
    VALUES (?, ?, ?, ?)
    ON CONFLICT(account) DO UPDATE SET token=excluded.token, latitude=excluded.latitude, longitude=excluded.longitude;
    `
	_, err := db.Exec(insertSQL, account, token, latitude, longitude)
	return err
}

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
