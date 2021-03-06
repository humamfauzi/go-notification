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
type IColumnMatcher interface {
	ColumnMatcher (columnName string) interface{}
	GetAllColumn() []interface{}
}

func dynamicScan(selectColumn []string, model IColumnMatcher) []interface{} {
	if len(selectColumn) == 1 && selectColumn[0] == "*" {
		return model.GetAllColumn()
	}
	scanArray := make([]interface{}, len(selectColumn))
	for index, v := range selectColumn {
		scanArray[index] = model.ColumnMatcher(v)
	}
	return scanArray
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
		for j := 0; j < len(pairs[i]); j++ {
			finalQuery += fmt.Sprintf(" %s ", pairs[i][j])
		}
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
	Token string `json:"token"`
}

func (up UserProfile) GetFilledKey() []string {
	nonEmpty := []string{}
	if up.Email != "" {
		nonEmpty = append(nonEmpty, "email")
	}
	if up.Password != "" {
		nonEmpty = append(nonEmpty, "password")
	}
	if up.Token != "" {
		nonEmpty = append(nonEmpty, "Token")
	}
	return nonEmpty
}

func (up UserProfile) ToStringJSON() string {
	result, _ := json.Marshal(up)
	return string(result)
}

func (up *UserProfile) ColumnMatcher(columnName string) interface{} {
	switch columnName {
	case "id":
		return &up.Id
	case "email":
		return &up.Email
	case "password":
		return &up.Password
	default:
		return nil
	}
}

func (up *UserProfile) GetAllColumn() []interface{} {
	return []interface{}{
		&up.Id,
		&up.Email,
		&up.Password,
	}
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
	if err := up.Scan(rows, selectColumn); err != nil {
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
	if err := up.Scan(rows, selectColumn); err != nil {
		return err
	}
	return nil
}
 
func (up *UserProfile) Scan(rows RowsScan, selectRows []string) error {
	defer rows.Close()
	count := 0
	scanArray := dynamicScan(selectRows, up)
	for rows.Next(){	
		if err := rows.Scan(scanArray...); err != nil {
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
	return fmt.Sprintf("('%s','%s', '%s')", up.Id, up.Email, up.Password)
}

func (up UserProfile) Insert(tx ITransaction) (int64, error) {
	path := "users.create"
	lastInsertId, err := WriteToDB(tx, path, up.InsertFormat())
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (up UserProfile) UpdateFormat(updateables []string) string {
	baseQuery := ""
	for _, column := range updateables {
		switch column {
		case "email":
			baseQuery += fmt.Sprintf("email = '%s',", up.Email)
		case "token":
			baseQuery += fmt.Sprintf("token = '%s',", up.Token)
		case "passowrd":
			baseQuery += fmt.Sprintf("password = '%s',", up.Password)
		default:
			baseQuery += ""
		}
	}
	baseQuery = strings.TrimSuffix(baseQuery, ",")
	return baseQuery
}

func (up UserProfile) Update(tx ITransaction, updateables []string) (int64, error) {
	path := "users.update"
	lastInsertId, err := WriteToDB(tx, path, up.UpdateFormat(updateables), up.Id)
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
	Id int `json:"id"`
	UserId string `json:"user_id"`
	Title string `json:"title"`
	Desc string `json:"description`
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

func (t *Topic) ColumnMatcher(column string) interface{} {
	switch column {
	case "id":
		return &t.Id
	case "user_id":
		return &t.UserId
	case "desc":
		return &t.Desc
	case "title":
		return &t.Title
	default:
		return nil
	}
}

func (t *Topic) GetAllColumn() []interface{} {
	return []interface{}{
		&t.Id,
		&t.UserId,
		&t.Desc,
		&t.Title,
	}
}

func (t Topics) Insert(tx ITransaction) (int64, error) {
	path := "topics.insert"
	lastInsertId, err := WriteToDB(tx, path, t.InsertFormat())
	if err != nil {
		return int64(0), err
	}
	return lastInsertId, nil
}

func (t *Topics) Get(tx ITransaction, selectColumn []string, wherePairs [][]string, afterWhere [][]string) error {
	path := "topics.get"
	rows, err := ReadFromDB(tx, path, selectColumn, wherePairs, afterWhere)
	if err != nil {
		return err
	}
	if err := t.Scan(rows, selectColumn); err != nil {
		return err
	}
	return nil
}

func (t *Topics) Scan(rows RowsScan, selectColumn []string) error {
	defer rows.Close()
	count := 0
	for rows.Next() {
		tb := &Topic{}
		scanArray := dynamicScan(selectColumn, tb)
		if err := rows.Scan(scanArray...); err != nil {
			return err
		}
		(*t) = append(*t, *tb)
		count++
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ------- SUBSCRIBER MODEL FUNCTION --------- //
type Subscriber struct {
	Id int `json:"id"`
	TopicId int `json:"topic_id"`
	UserId string `json:"user_id"`
}

func (s Subscriber) InsertFormat() string {
	return fmt.Sprintf("(%d,'%s')", s.TopicId, s.UserId)
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

func (s *Subscriber) ColumnMatcher(columnName string) interface{} {
	switch columnName {
	case "id":
		return &s.Id
	case "topic_id":
		return &s.TopicId
	case "user_id":
		return &s.UserId
	default:
		return nil
	}
}

func (s *Subscriber) GetAllColumn() []interface{} {
	return []interface{} {
		&s.Id,
		&s.TopicId,
		&s.UserId,
	}
}

type Subscribers []Subscriber

func (s *Subscribers) Get(tx ITransaction, selectColumn []string, wherePairs [][]string) error {
	path := "subscribers.get"
	if len(selectColumn) == 0 {
		selectColumn = []string{"*"}
	}
	rows, err := ReadFromDB(tx, path, selectColumn, wherePairs)
	if err != nil {
		return err
	}

	if err := s.Scan(rows, selectColumn); err != nil {
		return err
	}
	return nil
}

func (s *Subscribers) Scan(rows RowsScan, selectColumn []string) error {
	defer rows.Close()
	count := 0
	for rows.Next() {
		subscriber := &Subscriber{}
		scanArray := dynamicScan(selectColumn, subscriber)
		if err := rows.Scan(scanArray...); err != nil {
			return err
		}
		(*s) = append(*s, *subscriber)
		count++
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}



// -------- NOTIFICATION MODEL FUNCTION --------- //
type Notification struct {
	Id int `json:"id"`
	UserId string `json:"user_id"`
	TopicId int `json:"topic_id"`
	Message string `json:"message"`
	IsRead bool `json:"is_read"`
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
	if err := n.Scan(rows, selectColumn); err != nil {
		return err
	}
	return nil
}

func (n *Notification) ColumnMatcher(column string) interface{} {
	switch column {
	case "id":
		return &n.Id
	case "user_id":
		return &n.UserId
	case "topic_id":
		return &n.TopicId
	case "message":
		return &n.Message
	case "is_read":
		return &n.IsRead
	default:
		return nil
	}
}

func (n *Notification) GetAllColumn() []interface{} {
	return []interface{}{
		&n.Id,
		&n.UserId,
		&n.TopicId,
		&n.Message,
		&n.IsRead,
	}
}

func (n *Notification) Scan(rows RowsScan, selectRows []string) error {
	defer rows.Close()
	count := 0
	scanArray := dynamicScan(selectRows, n)
	for rows.Next() {
		if err := rows.Scan(scanArray...); err != nil {
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

func (n *Notifications) Get(tx ITransaction, selectColumn []string, wherePairs [][]string) error {
	path := "notifications.get"
	if len(selectColumn) == 0 {
		selectColumn = []string{"*"}
	}
	rows, err := ReadFromDB(tx, path, selectColumn, wherePairs)
	if err != nil {
		return err
	}
	if err := n.Scan(rows, selectColumn); err != nil {
		return err
	}
	return nil
}

func (n *Notifications) Scan(rows RowsScan, selectColumn []string) error {
	defer rows.Close()
	count := 0
	for rows.Next() {
		notif := &Notification{}
		scanArray := dynamicScan(selectColumn, notif)
		if err := rows.Scan(scanArray...); err != nil {
			return err
		}
		(*n) = append(*n, *notif)
		count++
	}
	if count == 0 {
		return sql.ErrNoRows
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