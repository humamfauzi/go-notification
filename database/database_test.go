package database

import (
	"testing"
	"database/sql"
)

func TestConnectionDatabase(t *testing.T) {
	connProp := MysqlDatabaseAccess{
		Username: "root",
		Password: "",
		Protocol: "tcp",
		Address: "localhost",
		DBName: "try1",
	}
	connDB, err := ConnectDatabase(connProp)
	if err != nil {
		t.Fatalf("Cannot Connect to DB")
	}
	err = connDB.Ping()
	if err != nil {
		t.Fatalf("Failed to Ping")
	}
	db = connDB
}

func TestGetUser(t *testing.T) {
	userId := "user/2020/05c82a88/b66e"
	result, err := GetUser(userId)
	if err != nil {
		
	}
	want := "new@user.com"
	yield := result.Email
	if want != yield {
		t.Fatalf("want %v yield %v", want, yield)
	}
}

func TestInsertUserEmailAndId(t *testing.T) {
	userId := "user/2021/dsa93d/s2se"
	email := "asdwer@alsdj.com"
	userProfile := UserProfile{
		Id: userId,
		Email: email,
	}
	if ok := InsertUserEmailAndId(userProfile); !ok {
		t.Fatalf("Error inserting")
	}

	result, err := GetUser(userId)
	if err != nil {
		t.Fatalf("Failed to get user profile %v", err)
	}
	if userProfile.Email != result.Email {
		t.Fatalf("Want %v yield %v", userProfile.Email, result.Email)
	}	
}

func TestUpdateUserEmail(t *testing.T) {
	userId := "user/2021/dsa93d/s2se"
	email := "asd@asd.com"
	userProfile := UserProfile{
		Email: email,
		Id: userId,
	}
	if ok := UpdateUserEmail(userProfile); !ok {
		t.Fatalf("Failed to udpate")
	}
	result, err := GetUser(userId)
	if err != nil {
		t.Fatalf("Failed to get user profile %v", err)
	}
	if userProfile.Email != result.Email {
		t.Fatalf("Want %v yield %v", userProfile.Email, result.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	userId := "user/2021/dsa93d/s2se"
	userProfile := UserProfile{
		Id: userId,
	}
	if ok := DeleteUser(userProfile); !ok {
		t.Fatalf("Failed to delete")
	}
	result, err := GetUser(userId)
	if err != sql.ErrNoRows {
		t.Fatalf("Failed to get user profile %v", err)
	}
	if result.Id != "" {
		t.Fatalf("Should be empty")
	} 
}