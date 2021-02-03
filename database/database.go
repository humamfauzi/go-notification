package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strings"
	"strconv"
	"errors"
	"context"
)

var (
	db *sql.DB
	queryMap QueryMap
)

type QueryMap map[string]map[string]interface{}

func (qm QueryMap) GetQuery(path string) (string, error) {
	stringSplit := strings.Split(path, ".")
	result, ok := qm[stringSplit[0]][stringSplit[1]].(string)
	if !ok {
		return "", errors.New("QUERY MAP PATH NOT FOUND")
	}
	return result, nil
}

type MysqlDatabaseAccess struct {
	Username string
	Password string
	Protocol string
	Address string
	DBName string
}

func (mda MysqlDatabaseAccess) ConnectDatabase() (*sql.DB, error) {
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

type ITransaction interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func CreateTransaction(dbFunction func(tx *sql.Tx) error) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := dbFunction(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func ConvertJsonToQueryMap(dir string) error {
	body, err := ioutil.ReadFile(dir)
	if err != nil {
		return err
	}
	qm := make(QueryMap)
	err = json.Unmarshal(body, &qm)
	if err != nil {
		return err
	}
	queryMap = qm
	return nil
}

func WriteToDB(tx ITransaction, path string, input ...interface{}) error {
	query, err := Query(path, input...)
	if err != nil {
		return err
	}
	fmt.Println(query)
	write, err := tx.Query(query)
	if err != nil {
		return err
	}
	write.Close()
	return nil
}

func Query(path string, input ...interface{}) (string, error) {
	formatQuery, err := queryMap.GetQuery(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(formatQuery, input...), nil
}

func createColumnValuePairing(pairs [][]string) string {
	if len(pairs) == 0 {
		return ""
	}
	finalQuery := " WHERE "
	for i := 0; i < len(pairs); i++ {
		if len(pairs[i]) == 1 {
			finalQuery += fmt.Sprintf(" %s ", pairs[i])
		} else if len(pairs[i]) == 3 {
			finalQuery += fmt.Sprintf(" %s %s '%s' ", pairs[i][0], pairs[i][1], pairs[i][2])
		}
	}
	return finalQuery
}

func ReadFromDB(tx ITransaction, path string, selectColumn []string, whereColumn [][]string) (RowsScan, error) {
	column := strings.Join(selectColumn, ",")
	whereQuery := createColumnValuePairing(whereColumn)
	query, err := Query(path, column, whereQuery)
	fmt.Println(query)
	if err != nil {
		return &sql.Rows{}, err
	}
	rows, err := tx.Query(query)
	if err != nil {
		return &sql.Rows{}, err
	}
	return rows, nil
}

func CloseConnection() {
	db.Close()
}

// ------- USER MODEL FUNCTION --------- //
type UserProfile struct {
	Email string
	Id string
}

type RowsScan interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
}

func (up *UserProfile) Get(tx ITransaction) error {
	path := "users.get"
	selectColumn := []string{"email", "id"}
	wherePairs := [][]string{
		[]string{
			"id", "=", up.Id,
		},
	}
	rows, err := ReadFromDB(tx, path, selectColumn, wherePairs)
	if err != nil {
		return err
	}
	if err := up.Scan(rows); err != nil {
		return err
	}
	return nil
}

func (up *UserProfile) Scan(rows RowsScan) error {
	defer rows.Close()
	count := 0
	for rows.Next(){	
		if err := rows.Scan(&up.Email, &up.Id); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (up UserProfile) InsertFormat() string {
	return fmt.Sprintf("('%s','%s')", up.Id, up.Email)
}

func (up UserProfile) Insert(tx ITransaction) error {
	path := "users.create"
	if err := WriteToDB(tx, path, up.InsertFormat()); err != nil {
		return err
	}
	return nil
}

func (up UserProfile) UpdateFormat() string {
	return fmt.Sprintf("'%s'", up.Email)
}

func (up UserProfile) Update(tx ITransaction) error {
	path := "users.update"
	if err := WriteToDB(tx, path, up.UpdateFormat(), up.Id); err != nil {
		return err
	}
	return nil
}

func (up UserProfile) DeleteFormat() string {
	return fmt.Sprintf("'%s'", up.Id)
}

func (up UserProfile) Delete(tx ITransaction) error {
	path := "users.delete"
	if err := WriteToDB(tx, path, up.DeleteFormat()); err != nil {
		return err
	}
	return nil
}

type UserProfiles []UserProfile

func (up UserProfiles) BulkCreate(tx ITransaction) error {
	for i:=0; i < len(up); i++ {
		err := up[i].Insert(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (up UserProfiles) BulkDelete(tx ITransaction) error {
	for i:=0; i < len(up); i++ {
		err := up[i].Delete(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

// ------- TOPIC MODEL FUNCTION --------- //
type Topic struct {
	Id int
	UserId string
	Title string
	Desc string
}

func (t Topic) InsertFormat() string {
	return fmt.Sprintf("('%s','%s','%s')", t.UserId, t.Title, t.Desc)
}

func (t Topic) Insert(tx ITransaction) error {
	path := "topic.insert"
	if err := WriteToDB(tx, path, t.InsertFormat()); err != nil {
		return err
	}
	return nil
}

func (t Topic) UpdateFormat() string {
	return fmt.Sprintf("(%s,%s)", t.Title, t.Desc)
}

func (t Topic) Update(tx ITransaction) error {
	path := "topic.update"
	if err := WriteToDB(tx, path, t.UpdateFormat); err != nil {
		return err
	}
	return nil
}

func (t Topic) Delete(tx ITransaction) error {
	path := "topic.delete"
	if err := WriteToDB(tx, path, t.Id); err != nil {
		return err
	}
	return nil
}

// ------- SUBSCRIBER MODEL FUNCTION --------- //
type Subscriber struct {
	Id int
	TopicId int
	UserId string
}
func (s Subscriber) CreateFormat() string {
	return fmt.Sprintf("(%d,%s)", s.TopicId, s.UserId)
}

func (s Subscriber) Create(tx ITransaction) error {
	path:= "subscriber.create"
	if err := WriteToDB(tx, path, s.CreateFormat()); err != nil {
		return err
	}
	return nil
}

func (s Subscriber) DeleteFormat() string {
	return fmt.Sprintf("%d", s.Id)
}

func (s Subscriber) Delete(tx ITransaction) error {
	path := "subscriber.delete"
	if err := WriteToDB(tx, path, s.DeleteFormat()); err != nil {
		return err
	}
	return nil
}

// -------- NOTIFICATION MODEL FUNCTION --------- //
type Notification struct {
	Id int
	UserId string
	TopicId int
	Message string
	IsRead bool
}

func (n Notification) InsertFormat() string {
	return fmt.Sprintf("('%s',%d,'%s')", n.UserId, n.TopicId, n.Message)
}

func (n Notification) Insert(tx ITransaction) error {
	path := "notification.insertNotification"
	if err := WriteToDB(tx, path, n.InsertFormat()); err != nil {
		return err
	}
	return nil
}

func (n *Notification) Get(tx ITransaction) error {
	path := "notification.get"
	selectColumn := []string{"id", "user_id", "topic_id", "message", "is_read"}
	wherePairs := [][]string{
		[]string{
			"id", "=", fmt.Sprintf("%d", n.Id),
		},
	}
	rows, err := ReadFromDB(tx, path, selectColumn, wherePairs)
	if err != nil {
		return err
	}
	if err := n.Scan(rows); err != nil {
		return err
	}
	return nil
}

func (n *Notification) Scan(rows RowsScan) error {
	defer rows.Close()
	count := 0
	for rows.Next() {
		if err := rows.Scan(&n.Id, &n.UserId, &n.TopicId, &n.Message, &n.IsRead); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type Notifications []Notification

func (n Notifications) InsertFormat() string {
	finalQuery := make([]string, len(n))
	for i:=0; i < len(n); i++ {
		finalQuery[i] = n[i].InsertFormat()
	}
	return strings.Join(finalQuery, ",")
}

func (n Notifications) Insert(tx ITransaction) error {
	path := "notification.bulkInsertNotification"
	if err := WriteToDB(tx, path, n.InsertFormat()); err != nil {
		return err
	}
	return nil
}

func (n Notifications) ComposeInputBulkFormat() string {
	finalFormat := make([]string, len(n))
	baseFormat := "('%s',%v,'%s')"
	for i:=0; i < len(n); i++ {
		finalFormat[i] = fmt.Sprintf(baseFormat, n[i].UserId, n[i].TopicId, n[i].Message)
	}
	return strings.Join(finalFormat, ",")
}

func (n Notifications) ComposeIdBulkFormat() string {
	finalFormat := make([]string, len(n))
	for i:=0; i < len(n); i++ {
		finalFormat[i] = strconv.Itoa(n[i].Id)
	}
	return "(" + strings.Join(finalFormat, ",") + ")"
}

func (n Notifications) UpdateReadNotification(tx ITransaction) error {
	path := "notification.updateReadNotification"
	if err := WriteToDB(tx, path, n.ComposeIdBulkFormat()); err != nil {
		return err
	}
	return nil
}