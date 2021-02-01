package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	// "fmt"
)

var (
	db *sql.DB
	queryMap QueryMap
)

type MysqlDatabaseAccess struct {
	Username string
	Password string
	Protocol string
	Address string
	DBName string
}

type QueryMap map[string]string

func ConnectDatabase(mda MysqlDatabaseAccess) (*sql.DB, error) {
	composed := mda.Username + ":"
	composed += mda.Password + "@"
	composed += mda.Protocol + "("
	composed += mda.Address + ")/"
	composed += mda.DBName
	db, err := sql.Open("mysql", composed)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CloseConnection() {
	db.Close()
}

type UserProfile struct {
	Email string
	Id string
}

func GetUser(userId string) (UserProfile, error) {
	var userProfile UserProfile
	row := db.QueryRow("SELECT id, email FROM users WHERE id=?", userId)
	switch err := row.Scan(&userProfile.Id, &userProfile.Email); err {
	case nil:
		return userProfile, nil
	default:
		return userProfile, err
	}
}

func InsertUserEmailAndId(userProfile UserProfile) bool {
	insert, err := db.Query("INSERT INTO users (id, email) VALUES (?, ?)", userProfile.Id, userProfile.Email)
	if err != nil {
		return false
	}
	insert.Close()
	return true
}

func UpdateUserEmail(userProfile UserProfile) bool {
	update, err := db.Query("UPDATE users SET email = ? WHERE id = ?", userProfile.Email, userProfile.Id)
	if err != nil {
		return false
	}
	update.Close()
	return true
}

func DeleteUser(userProfile UserProfile) bool {
	delete, err := db.Query("DELETE FROM users WHERE id = ?", userProfile.Id) 
	if err != nil {
		return false
	}
	delete.Close()
	return true
}

