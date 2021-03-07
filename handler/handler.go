package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"

	dba "github.com/humamfauzi/go-notification/database"
	"github.com/humamfauzi/go-notification/utils"
	"github.com/humamfauzi/go-notification/auth"
)

const (
	QUERY_MAP_RELATIVE_LOCATION = "../database/queryMap.json"
)

var (
	dbConn dba.ITransactionSQL
)

/**
	This interface ensure that any database that connected to handler have
	ITransaction interface which is ability to Query and Exec a statement

	Different database service need to be wrapped in such that implement
	both Query and Exec. This is ensure we can connect other than
	MySQL when needed arises.

	This also implicitly tells that handler does not care what database it impelemented
	as long as capability to excute a query

*/
type DbConnection interface {
	ConnectDatabase() (dba.ITransactionSQL, error)
}

func ConnectToDatabase(dbProfile DbConnection) {
	connDB, err := dbProfile.ConnectDatabase()
	dba.ConvertJsonToQueryMap(QUERY_MAP_RELATIVE_LOCATION)
	if err != nil {
		panic(err)
	}
	dbConn = connDB
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "HomePage")
	return
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("ACCESSED", r.Header.Get("User-Agent"))
		next.ServeHTTP(w, r)
	})
}

func getUserProfileFromAuth(accessToken string) (dba.UserProfile, error) {
	userProfile := dba.UserProfile{}
	wherePairs := [][]string{
		[]string{
			"token", "=", accessToken,
		},
	}
	if err := userProfile.Find(dbConn, []string{"id"}, wherePairs); err != nil {
		return userProfile, err
	}
	return userProfile, nil
}

func getRequesterProfile(r *http.Request) (dba.UserProfile, error) {
	requester := r.Header.Get("requesterProfile")
	userProfile := dba.UserProfile{}
	if len(requester) == 0 {
		return userProfile, errors.New("Unknown Requester")
	}
	identity := []byte(requester)	
	if err := json.Unmarshal(identity, &userProfile); err != nil {
		return userProfile, errors.New("Unknown Requester")
	}
	return userProfile, nil
}

func TokenCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticationToken := r.Header.Get("Authentication")
		if len(authenticationToken) == 0 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		accessToken, ok := auth.VerifyToken(authenticationToken, auth.KeyFunction);
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		userProfile, err := getUserProfileFromAuth(accessToken)
		if err != nil {
			WriteReply(int(http.StatusBadRequest), false, "Cannot find matched Token", w)
			return
		}
		r.Header.Add("requesterProfile", userProfile.ToStringJSON())
		next.ServeHTTP(w, r)
	})
}

type HandlerReply struct {
	Code int `json:"code"`
	Success bool `json:"sucess"`
	Message interface{} `json:"message"`
}

func WriteReply(code int, success bool, message interface{}, w http.ResponseWriter) {
	hr := HandlerReply{
		Code: code,
		Success: success,
		Message: message,
	}
	reply, _ := json.Marshal(hr)
	w.WriteHeader(code)
	w.Write(reply)
}


func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}
	userProfile := dba.UserProfile{}
	if err := json.Unmarshal(body, &userProfile); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	
	storedPassword, err := auth.BcryptConvertTo(userProfile.Email, userProfile.Password)
	if err != nil {
		WriteReply(int(http.StatusInternalServerError), false, "Cannot use auth conversion", w)
		return
	}

	userProfile.Id = utils.RandomStringId("user", 10)
	userProfile.Password = storedPassword
	if _, err := userProfile.Insert(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, fmt.Sprintf("Cannot Write Payload %v", err), w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

type LoginOps struct {}

func (lo LoginOps) searchUserByEmailAndCheckPassword(email, password string) (dba.UserProfile, error) {
	profileFromDB := dba.UserProfile{}
	selectCols := []string{"id", "email", "password"}
	wherePairs := [][]string{
		[]string{
			"email", "=", email,
		},
	}
	if err := profileFromDB.Find(dbConn, selectCols, wherePairs); err != nil {
		return profileFromDB, err
	}
	storedPassword := []byte(profileFromDB.Password)
	requesterPassword := auth.ComposeBcryptPassword(email, password)

	if ok := auth.BcryptCheck(storedPassword, requesterPassword); !ok {
		return profileFromDB, errors.New("Password Unmatched")
	}
	return profileFromDB, nil
}

func (lo LoginOps) generateJWT(token string) (string, error) {
	mapClaims := make(map[string]interface{})
	mapClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	mapClaims["access_token"] = token
	return auth.CreateToken(mapClaims, auth.GetAuthSecret())
}

func (lo LoginOps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}

	userProfile := dba.UserProfile{}
	if err := json.Unmarshal(body, &userProfile); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	storedUserProfile, err :=lo.searchUserByEmailAndCheckPassword(userProfile.Email, userProfile.Password); 
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, fmt.Sprintf("Wrong password %v", err), w)
		return
	}

	accessToken := utils.RandomStringId("accessToken", 64)
	token, err := lo.generateJWT(accessToken)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Token Generation Error", w)
		return
	}
	storedUserProfile.Token = accessToken
	storedUserProfile.Update(dbConn, []string{"token"})
	reply := struct {
		Token string `json:"token"`
	}{ token }
	WriteReply(int(http.StatusOK), true, reply, w)
	return
}

type CheckLogin struct {}

func (cl CheckLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requester := r.Header.Get("requesterProfile")
	if len(requester) == 0 {
		WriteReply(int(http.StatusBadRequest), false, "Unknown Requester", w)
		return
	}
	WriteReply(int(http.StatusOK), true, "Login Verfied", w)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}
	userProfile := dba.UserProfile{
		Id: vars["id"],
	}
	if err := json.Unmarshal(body, &userProfile); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	updateables := userProfile.GetFilledKey()
	if _, err := userProfile.Update(dbConn, updateables); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, err := getRequesterProfile(r)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot identify requester", w)
	}
	if _, err := userProfile.Delete(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

func CreateTopicHandler(w http.ResponseWriter,r *http.Request) {
	userProfile, err := getRequesterProfile(r)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot identify requester", w)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}
	topicProfile := dba.Topic{}
	if err := json.Unmarshal(body, &topicProfile); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	topicProfile.UserId = userProfile.Id
	if _, err := topicProfile.Insert(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, err := getRequesterProfile(r)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot identify requester", w)
	}
	selectColumn := []string{"*"}
	wherePairs := [][]string{
		[]string {
			"user_id", "=", userProfile.Id,
		},
	}

	topicProfiles := dba.Topics{}
	if err := topicProfiles.Get(dbConn, selectColumn, wherePairs, [][]string{}); err != nil {
		fmt.Println(err)
		WriteReply(int(http.StatusBadRequest), false, "Cannot Get Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, topicProfiles, w)
	return
}

func CreateSubscribeHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, err := getRequesterProfile(r)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot identify requester", w)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}
	subscriberProfile := dba.Subscriber{}
	if err := json.Unmarshal(body, &subscriberProfile); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	subscriberProfile.UserId = userProfile.Id
	if _, err := subscriberProfile.Insert(dbConn); err != nil {
		fmt.Println(err)
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

type CreateNotification struct {}

func (cn CreateNotification) GetAllSubscribers(topicId int) ([]string, error) {
	users := dba.Subscribers{}
	selectColumn := []string{"user_id"}
	wherePairs := [][]string{
		[]string{
			"topic_id", "=", strconv.Itoa(topicId),
		},
	}
	if err := users.Get(dbConn, selectColumn, wherePairs); err != nil {
		return []string{}, err
	}
	userId := make([]string, len(users))
	for i:=0; i < len(users); i++ {
		userId[i] = users[i].UserId
	}
	return userId, nil
}

func (cn CreateNotification) ComposeNotification(users []string, topicId int, message string) dba.Notifications {
	notificationList := make([]dba.Notification, len(users))
	for i := 0; i < len(users); i++ {
		notificationList[i].UserId = users[i]
		notificationList[i].Message = message
		notificationList[i].TopicId = topicId
	}
	return dba.Notifications(notificationList)
}

func (cn CreateNotification) IsTopicBelongToUser(userId string, topicId int) bool {
	selectColumn := []string{"user_id"}
	wherePairs := [][]string{
		[]string{
			"user_id", "=", userId,
		},
		[]string{
			"topic_id", "=", strconv.Itoa(topicId),
		},
	}
	afterWhere := [][]string{
		[]string{
			"order by", "id", "desc",
		},
		[]string{
			"limit", "1",
		},
	}
	topics := dba.Topics{}
	if err := topics.Get(dbConn, selectColumn, wherePairs, afterWhere); err != nil {
		return false
	}
	if len(topics) < 1 {
		return false
	}
	return true
}

func (cn CreateNotification) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}
	request:= dba.Notification{}
	if err := json.Unmarshal(body, &request); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	
	users, err := cn.GetAllSubscribers(request.TopicId)
	if err != nil {
		fmt.Println(err)
		WriteReply(int(http.StatusBadRequest), false, "Cannot Get All Subscriber", w)
		return
	}

	notifications := cn.ComposeNotification(users, request.TopicId, request.Message)
	if _, err := notifications.Insert(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

func GetNotificationHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, err := getRequesterProfile(r)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot identify requester", w)
		return
	}
	notifications := dba.Notifications{}
	selectColumn := []string{"*"}
	wherePairs := [][]string{
		[]string{"user_id", "=", userProfile.Id},
	}
	if err := notifications.Get(dbConn, selectColumn, wherePairs); err != nil {
		fmt.Println(err)
		WriteReply(int(http.StatusBadRequest), false, "Cannot Get Payload", w)
		return
	}
	reply, err := json.Marshal(notifications)
	if err != nil {
		WriteReply(int(http.StatusInternalServerError), false, "Cannot Wrap Result", w)
		return
	}
	WriteReply(int(http.StatusOK), true, reply, w)
	return
}
