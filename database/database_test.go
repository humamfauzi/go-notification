package database

import (
	"testing"
	"database/sql"
	"reflect"
)

var (
	db *sql.DB
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

func TestGetUser(t *testing.T) {
	userProfile := UserProfile{
		Id: "user/2020/05c82a88/b66e",
	}
	err := userProfile.Get(db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := "new@user.com"
	yield := userProfile.Email
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

	if len(getTopics) != 2 {
		t.Fatalf("topics len should equal to %d, %v", 2, getTopics)
	}
}

func TestUpdateNotificationIsRead(t *testing.T) {
	user1 := UserProfile{
		Id: "user/1",
		Email: "asdf@adsf.com",
	}
	user2 := UserProfile{
		Id: "user/2",
		Email: "dddd@gmz.com",
	}
	user3 := UserProfile{
		Id: "user/3",
		Email: "aaaa@gmz.com",
	}
	topic := Topic{
		Id: 1,
		UserId: "user/3",
		Title: "Hello",
		Desc: "try1",
	}
	insertArray := Notifications{
		Notification{
			Id: 1,
			UserId: "user/1",
			TopicId: 1,
			Message: "Hello",
		},
		Notification{
			Id: 2,
			UserId: "user/2",
			TopicId: 1,
			Message: "Hello",
		},
	}
	if err := CreateTransaction(func(tx *sql.Tx) error {
		_, err := insertArray.Delete(tx)
		if err != nil {
			return err
		}
		_, err = topic.Delete(tx)
		if err != nil {
			return err
		}
		_, err = user1.Delete(tx)
		if err != nil {
			return err
		}
		_, err = user2.Delete(tx)
		if err != nil {
			return err
		}
		_, err = user3.Delete(tx)
		if err != nil {
			return err
		}
	_, err = user1.Insert(tx)
		if err != nil {
			return err
		}
		_, err = user2.Insert(tx)
		if err != nil {
			return err
		}
		_, err = user3.Insert(tx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("%v", err)
	}

	
	if _, err := topic.Insert(db); err != nil {
		t.Fatalf("%v", err)
	}
	if _, err := insertArray.Insert(db); err != nil {
		t.Fatalf("%v", err)
	}
	
	if _, err := insertArray.UpdateReadNotification(db); err != nil {
		t.Fatalf("Failed to update notification")
	}
	note := Notification{
		Id: 1,
	}
	if err := note.Get(db); err != nil {
		t.Fatalf("Failed to get notification %v", err)
	}
	if !note.IsRead {
		t.Fatalf("want %t yield %t", true, note.IsRead)
	}
}