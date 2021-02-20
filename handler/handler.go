package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"io/ioutil"
	"log"
	"net/http"
	
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

func getUserProfileFromAuth(authenticationToken string) (dba.UserProfile, error) {
	token := auth.ParseBearer(authenticationToken)
	userProfile := dba.UserProfile{}
	wherePairs := [][]string{
		[]string{
			"token", "=", token,
		},
	}
	if err := userProfile.Find(dbConn, []string{"id"}, wherePairs); err != nil {
		return userProfile, err
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
		if ok := auth.VerifyToken(authenticationToken, auth.KeyFunction); !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		userProfile, err := getUserProfileFromAuth(authenticationToken)
		if err != nil {
			fmt.Println(err)
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

func (lo LoginOps) generateToken() (string, error) {
	mapClaims := make(map[string]interface{})
	mapClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
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

	token, err := lo.generateToken()
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Token Generation Error", w)
		return
	}
	storedUserProfile.Token = token
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
	if _, err := userProfile.Delete(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

func CreateTopicHandler(w http.ResponseWriter,r *http.Request) {
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
	if _, err := topicProfile.Insert(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}

func GetTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicProfiles := dba.Topics{}
	if err := topicProfiles.Get(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Get Payload", w)
		return
	}
	reply, err := json.Marshal(topicProfiles)
	if err != nil {
		WriteReply(int(http.StatusInternalServerError), false, "Cannot Wrap Result", w)
		return
	}
	WriteReply(int(http.StatusOK), true, reply, w)
	return
}
func CreateSubscribeHandler(w http.ResponseWriter, r *http.Request) {
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
	if _, err := subscriberProfile.Insert(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}
func CreateNotificationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Read Payload", w)
		return
	}
	notifications := dba.Notifications{}
	if err := json.Unmarshal(body, &notifications); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Parse Payload", w)
		return
	}
	if _, err := notifications.Insert(dbConn); err != nil {
		WriteReply(int(http.StatusBadRequest), false, "Cannot Write Payload", w)
		return
	}
	WriteReply(int(http.StatusOK), true, nil, w)
	return
}
// func GetNotificationHandler(w http.ResponseWriter, r *http.Request) {
// 	notifications := dba.Notifications{}
// 	if err := notifications.Get(dbConn); err != nil {
// 		WriteReply(int(http.StatusBadRequest), false, "Cannot Get Payload", w)
// 		return
// 	}
// 	reply, err := json.Marshal(notifications)
// 	if err != nil {
// 		WriteReply(int(http.StatusInternalServerError), false, "Cannot Wrap Result", w)
// 		return
// 	}
// 	WriteReply(int(http.StatusOK), true, reply, w)
// 	return
// }
