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
)

var (
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

func (mda MysqlDatabaseAccess) ConnectDatabase() (ITransactionSQL, error) {
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
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type ITransactionSQL interface {
	Begin() (*sql.Tx, error)
	Ping() error
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func CreateSQLTransaction(dbEngine ITransactionSQL, dbFunction func(tx *sql.Tx) error) error {
	tx, err := dbEngine.Begin()
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

func WriteToDB(tx ITransaction, path string, input ...interface{}) (int64, error) {
	query, err := Query(path, input...)
	if err != nil {
		return 0, err
	}
	fmt.Println(query)
	write, err := tx.Exec(query)
	if err != nil {
		return 0, err
	}
	lastInsertId, err := write.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertId, nil
}

func Query(path string, input ...interface{}) (string, error) {
	formatQuery, err := queryMap.GetQuery(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(formatQuery, input...), nil
}
func createAfterWherePair(pairs [][]string) string {
	if len(pairs) == 0 {
		return ""
	}
	finalQuery := ""
	for i := 0; i < len(pairs); i++ {
		finalQuery += fmt.Sprintf(" %s %s ", pairs[i][0], pairs[i][1])
	}
	return finalQuery
}

func createColumnValuePairing(pairs [][]string) string {
	if len(pairs) == 0 {
		return ""
	}
	finalQuery := " WHERE "
	for i := 0; i < len(pairs); i++ {
		if len(pairs[i]) == 1 {
			// for an or, and, parentheses
			finalQuery += fmt.Sprintf(" %s ", pairs[i])
		} else if len(pairs[i]) == 3 {
			// for a condilitionals e.g '=', '!=', 'is not null'
			finalQuery += fmt.Sprintf(" %s %s '%s' ", pairs[i][0], pairs[i][1], pairs[i][2])
		}
	}
	return finalQuery
}

/**
	Create an interface after path, currently can be inserted with
	selectColumn []string, whereColumn [][]string, afterWhere[][]string
	must be inserted in that order
*/
func ReadFromDB(tx ITransaction, path string, selectionInterface ...interface{}) (RowsScan, error) {
	var selectColumn []string
	var whereColumn, afterWhere [][]string
	switch len(selectionInterface) {
	case 1:
		selectColumn = selectionInterface[0].([]string)
	case 2:
		selectColumn = selectionInterface[0].([]string)
		whereColumn = selectionInterface[1].([][]string)
	case 3:
		selectColumn = selectionInterface[0].([]string)
		whereColumn = selectionInterface[1].([][]string)
		afterWhere = selectionInterface[2].([][]string)
	default:
		return &sql.Rows{}, errors.New("Query Requires Filter")

	}
	column := strings.Join(selectColumn, ",")
	whereQuery := createColumnValuePairing(whereColumn)
	afterWhereQuery := createAfterWherePair(afterWhere)
	whereQuery += afterWhereQuery
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

func CloseConnection(dbEngine ITransactionSQL) {
	dbEngine.Close()
}

type RowsScan interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
}

// ------- USER MODEL FUNCTION --------- //
type UserProfile struct {
	Email string `json:"email"`
	Id string `json:"id"`
	Password string `json:"password"`
}

func (up UserProfile) toStringJSON() string {
	result, _ := json.Marshal(up)
	return string(result)
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

func (up *UserProfile) Find(tx ITransaction, selectColumn []string, wherePairs[][]string) error {
	path := "users.find"
	if len(selectColumn) == 0 {
		selectColumn = []string{"*"}
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

func (up UserProfile) Insert(tx ITransaction) (int64, error) {
	path := "users.create"
	lastInsertId, err := WriteToDB(tx, path, up.InsertFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (up UserProfile) UpdateFormat() string {
	return fmt.Sprintf("'%s'", up.Email)
}

func (up UserProfile) Update(tx ITransaction) (int64, error) {
	path := "users.update"
	lastInsertId, err := WriteToDB(tx, path, up.UpdateFormat(), up.Id)
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (up UserProfile) DeleteFormat() string {
	return fmt.Sprintf("'%s'", up.Id)
}

func (up UserProfile) Delete(tx ITransaction) (int64, error) {
	path := "users.delete"
	lastInsertId, err := WriteToDB(tx, path, up.DeleteFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

type UserProfiles []UserProfile

func (up UserProfiles) BulkInsert(tx ITransaction) (int64, error) {
	var lastInsertId int64
	lastInsertId = 0
	for i:=0; i < len(up); i++ {
		lastInsertId, err := up[i].Insert(tx)
		if err != nil {
			return lastInsertId, err
		}
	}
	return lastInsertId, nil
}

func (up UserProfiles) BulkDelete(tx ITransaction) error {
	for i:=0; i < len(up); i++ {
		_, err := up[i].Delete(tx)
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

func (t Topic) Insert(tx ITransaction) (int64, error) {
	path := "topic.insert"
	lastInsertId, err := WriteToDB(tx, path, t.InsertFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (t Topic) UpdateFormat() string {
	return fmt.Sprintf("(%s,%s)", t.Title, t.Desc)
}

func (t Topic) Update(tx ITransaction) (int64, error) {
	path := "topic.update"
	lastInsertId, err := WriteToDB(tx, path, t.UpdateFormat)
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (t Topic) Delete(tx ITransaction) (int64, error) {
	path := "topic.delete"
	lastInsertId, err := WriteToDB(tx, path, t.Id)
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

type Topics []Topic

func (t Topics) InsertFormat() string {
	finalQuery := make([]string, len(t))
	for i:=0; i < len(t); i++ {
		finalQuery[i] = t[i].InsertFormat()
	}
	return strings.Join(finalQuery, ",")
}

func (t Topics) Insert(tx ITransaction) (int64, error) {
	path := "topics.insert"
	lastInsertId, err := WriteToDB(tx, path, t.InsertFormat())
	if err != nil {
		return int64(0), err
	}
	return lastInsertId, nil
}

func (t *Topics) Get(tx ITransaction) error {
	path := "topics.get"
	selectColumn := []string{"id", "user_id", "title", "description"}
	wherePairs := [][]string{}
	afterWhere := [][]string{
		[]string{
			"limit", "10",
		},
	}
	rows, err := ReadFromDB(tx, path, selectColumn, wherePairs, afterWhere)
	if err != nil {
		return err
	}
	if err := t.Scan(rows); err != nil {
		return err
	}
	return nil
}

func (t *Topics) Scan(rows RowsScan) error {
	defer rows.Close()
	count := 0
	for rows.Next() {
		tb := Topic{}
		if err := rows.Scan(&tb.Id, &tb.UserId, &tb.Title, &tb.Desc); err != nil {
			return err
		}
		(*t) = append(*t, tb)
		count++
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ------- SUBSCRIBER MODEL FUNCTION --------- //
type Subscriber struct {
	Id int
	TopicId int
	UserId string
}
func (s Subscriber) InsertFormat() string {
	return fmt.Sprintf("(%d,%s)", s.TopicId, s.UserId)
}

func (s Subscriber) Insert(tx ITransaction) (int64, error) {
	path:= "subscriber.create"
	lastInsertId, err := WriteToDB(tx, path, s.InsertFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (s Subscriber) DeleteFormat() string {
	return fmt.Sprintf("%d", s.Id)
}

func (s Subscriber) Delete(tx ITransaction) (int64, error) {
	path := "subscriber.delete"
	lastInsertId, err := WriteToDB(tx, path, s.DeleteFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
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

func (n Notification) Insert(tx ITransaction) (int64, error) {
	path := "notification.insertNotification"
	lastInsertId, err := WriteToDB(tx, path, n.InsertFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
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

func (n Notifications) Insert(tx ITransaction) (int64, error) {
	path := "notification.bulkInsertNotification"
	lastInsertId, err := WriteToDB(tx, path, n.InsertFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
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

func (n Notifications) UpdateReadNotification(tx ITransaction) (int64, error) {
	path := "notifications.updateRead"
	lastInsertId, err := WriteToDB(tx, path, n.ComposeIdBulkFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (n Notifications) Delete(tx ITransaction) (int64, error) {
	path := "notifications.delete"
	lastInsertId, err := WriteToDB(tx, path, n.ComposeIdBulkFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}