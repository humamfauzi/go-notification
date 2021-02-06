package database

import (
	"testing"
	"database/sql"
	"reflect"
)

var (
	db ITransactionSQL
)

func TestConvertJsonToQueryMap(t *testing.T) {
	dir := "queryMap.json"
	if err := ConvertJsonToQueryMap(dir); err != nil {
		t.Fatalf("Failed to read query map %v", err)
	}
	qm := make(QueryMap)
	if reflect.TypeOf(queryMap) != reflect.TypeOf(qm) {
		t.Fatalf("Not a query map")
	}
}

func TestConnectionDatabase(t *testing.T) {
	connProp := MysqlDatabaseAccess{
		Username: "root",
		Password: "",
		Protocol: "tcp",
		Address: "localhost",
		DBName: "try1",
	}
	connDB, err := connProp.ConnectDatabase()
	if err != nil {
		t.Fatalf("Cannot Connect to DB")
	}
	err = connDB.Ping()
	if err != nil {
		t.Fatalf("Failed to Ping")
	}
	db = connDB
}

func TestInsertUserEmailAndId(t *testing.T) {
	userId := "user/2021/dsa93d/s2se"
	email := "asdwer@alsdj.com"
	userProfile := UserProfile{
		Id: userId,
		Email: email,
	}
	_, err := userProfile.Insert(db); 
	if err != nil {
		t.Fatalf("Failed to Insert User %v", err)
	}

	newUserProfile := UserProfile{
		Id: userId,
	}
	if err = newUserProfile.Get(db); err != nil {
		t.Fatalf("Failed to get user profile %v", err)
	}
	if userProfile.Email != newUserProfile.Email {
		t.Fatalf("Want %v yield %v", userProfile.Email, newUserProfile.Email)
	}	
}

func TestUpdateUserEmail(t *testing.T) {
	userId := "user/2021/dsa93d/s2se"
	email := "asd@asd.com"
	userProfile := UserProfile{
		Email: email,
		Id: userId,
	}
	_, err := userProfile.Update(db) 
	if err != nil {
		t.Fatalf("Failed to udpate %v", err)
	}
	newUserProfile := UserProfile{
		Id: userId,
	}
	err = newUserProfile.Get(db)
	if err != nil {
		t.Fatalf("Failed to get user profile %v", err)
	}
	if userProfile.Email != newUserProfile.Email {
		t.Fatalf("Want %v yield %v", userProfile.Email, newUserProfile.Email)
	}
}

func TestGetUser(t *testing.T) {
	userProfile := UserProfile{
		Id: "user/2021/dsa93d/s2se",
	}
	err := userProfile.Get(db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := "asd@asd.com"
	yield := userProfile.Email
	if want != yield {
		t.Fatalf("want %v yield %v", want, yield)
	}
}

func TestDeleteUser(t *testing.T) {
	userId := "user/2021/dsa93d/s2se"
	userProfile := UserProfile{
		Id: userId,
	}
	_, err := userProfile.Delete(db); 
	if err != nil {
		t.Fatalf("Failed to delete %v", err)
	}
	newUserProfile := UserProfile{
		Id: userId,
	}
	err = newUserProfile.Get(db)
	if err != sql.ErrNoRows {
		t.Fatalf("Failed to get user profile %v", err)
	}
	if newUserProfile.Email != "" {
		t.Fatalf("Should be empty")
	} 
}

func TestBulkFormat(t *testing.T) {
	insertArray := Notifications{
		Notification{
			Id: 23,
			UserId: "user/1",
			TopicId: 2,
			Message: "Hello",
		},
		Notification{
			Id: 24,
			UserId: "user/2",
			TopicId: 2,
			Message: "Hello",
		},
	}
	bulkFormat := insertArray.ComposeInputBulkFormat()
	want := `('user/1',2,'Hello'),('user/2',2,'Hello')`
	if want != bulkFormat {
		t.Fatalf("\nwant %v; \nget  %v", want, bulkFormat)
	}
	bulkFormat = insertArray.ComposeIdBulkFormat()
	want = `(23,24)`
	if want != bulkFormat {
		t.Fatalf("\nwant %v; \nget  %v", want, bulkFormat)
	}
}

func TestGetTopics(t *testing.T) {
	userId := "user/topic"
	userProfile := UserProfile{
		Id: userId,
		Email: "Topics@top.ics",
	}
	if _, err := userProfile.Insert(db); err != nil  {
		t.Logf("Failed to create user %v", err)
	}
	topics := Topics{
		Topic{
			UserId: userId,
			Title: "topic1",
		},
		Topic{
			UserId: userId,
			Title: "topic2",
		},
	}
	if _, err := topics.Insert(db); err != nil {
		t.Fatalf("Cannot insert topics %v", err)
	}

	getTopics := Topics{}
	if err := getTopics.Get(db); err != nil {
		t.Fatalf("Cannot get topics %v", err)
	}

	if len(getTopics) == 0 {
		t.Fatalf("topics len should equal to %d, %v", 2, getTopics)
	}
}