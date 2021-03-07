package handler

import (
	"testing"
	"strings"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	dba "github.com/humamfauzi/go-notification/database"
	"github.com/humamfauzi/go-notification/utils"
	"fmt"
	b64 "encoding/base64"
)

const (
	baseUrl = "http://example.com"
)

type handlerBuffer struct {
	funcStorage func(w http.ResponseWriter, r *http.Request)
}
func (hb *handlerBuffer) WrapFunction(originalFunction func(w http.ResponseWriter, r *http.Request)) {
	hb.funcStorage = originalFunction
}

func (hb handlerBuffer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hb.funcStorage(w, r)
}

func functionWrapper(handlerFunc func(w http.ResponseWriter, r *http.Request)) http.Handler {
	hb := handlerBuffer{}
	hb.WrapFunction(handlerFunc)
	return hb
}

func extractReply(w *httptest.ResponseRecorder) HandlerReply {
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)
	return reply
}

func createUserAndLogin(t *testing.T) string {
	randName := utils.RandomStringId("", 10)
	exampleJson := fmt.Sprintf(`{
		"email": "login%v@asd.asd",
		"password": "rahasia"
	}`, randName)
	jsonReader := strings.NewReader(exampleJson)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/users", jsonReader)
	w := httptest.NewRecorder()
	CreateUserHandler(w, req)

	jsonReader = strings.NewReader(exampleJson)
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/users/login", jsonReader) 
	w = httptest.NewRecorder()
	loginOps := LoginOps{}
	loginOps.ServeHTTP(w, req)
	
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)
	t.Logf("%v", reply.Message)
	token := reply.Message.(map[string]interface{})["token"].(string)
	return token
}

func TestConnectToDatabase(t *testing.T) {
	connProp := dba.MysqlDatabaseAccess{
		Username: "root",
		Password: "",
		Protocol: "tcp",
		Address: "localhost",
		DBName: "try1",
	}
	ConnectToDatabase(connProp)
}

func TestCreateUserHandler(t *testing.T) {
	exampleJson := `{
		"email": "asd@asd.asd",
		"password": "rahasia"
	}`
	jsonReader := strings.NewReader(exampleJson)
	handler := CreateUserHandler
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/users", jsonReader)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)
	t.Log(reply, string(body))
	if !reply.Success {
		t.Fatalf("Want true but have false")
	}
	if reply.Code != http.StatusOK {
		t.Fatalf("Want %v have %v", reply.Code, http.StatusOK)
	}
	return
}

func TestLogin(t *testing.T) {
	exampleJson := `{
		"email": "login@asd.asd",
		"password": "rahasia"
	}`
	jsonReader := strings.NewReader(exampleJson)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/users", jsonReader)
	w := httptest.NewRecorder()
	CreateUserHandler(w, req)

	jsonReader = strings.NewReader(exampleJson)
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/users/login", jsonReader) 
	w = httptest.NewRecorder()
	loginOps := LoginOps{}
	loginOps.ServeHTTP(w, req)
	
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)
	
	t.Log(reply, string(body))
	if !reply.Success {
		t.Fatalf("Want true but have false")
	}
	if reply.Code != http.StatusOK {
		t.Fatalf("Want %v have %v", reply.Code, http.StatusOK)
	}
	if reply.Message == nil {
		t.Fatalf("should contain the token")
	}
	return
}

func TestCheckLogin(t *testing.T) {
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader("")
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/users/check", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json",}
	w := httptest.NewRecorder()
	checkLogin := CheckLogin{}
	wrappedFunc := TokenCheckMiddleware(checkLogin)
	wrappedFunc.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}
}

func TestDeleteUserHandler(t *testing.T) {
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader("")
	req := httptest.NewRequest(http.MethodDelete, baseUrl + "/users/delete", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	fw := functionWrapper(DeleteUserHandler)
	wrappedFunc := TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}
	return
}
func TestCreateTopicHandler(t *testing.T) {
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader(`{
		"title": "example topic",
		"desc": "create topic test"
	}`)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	fw := functionWrapper(CreateTopicHandler)
	wrappedFunc := TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}
	return
}
func TestGetTopicHandler(t *testing.T) {
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader(`{
		"title": "example topic",
		"desc": "create topic test"
	}`)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	fw := functionWrapper(CreateTopicHandler)
	wrappedFunc := TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}
	
	jsonReader = strings.NewReader("")
	req = httptest.NewRequest(http.MethodGet, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w = httptest.NewRecorder()
	fw = functionWrapper(GetTopicHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	reply = HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}
	parsed := reply.Message.([]interface{})
	if len(parsed) == 0 {
		t.Fatalf("Should contain Topic")
	}
	return
}

func TestCreateSubscribeHandler(t *testing.T) {
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader(`{
		"title": "example topic",
		"desc": "create topic test"
	}`)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	fw := functionWrapper(CreateTopicHandler)
	wrappedFunc := TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}

	jsonReader = strings.NewReader(`{
		"title": "example topic",
		"desc": "create topic test for subscriber"
	}`)
	req = httptest.NewRequest(http.MethodGet, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(GetTopicHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("want 200 but get %v", reply.Code)
	}
	if reply.Message == nil {
		t.Fatalf("Message should not be nil")
	}

	parsed := reply.Message.([]interface{})
	choosenTopics := dba.Topic{}
	jsonResult, err := json.Marshal(parsed[0])
	if err != nil {
		t.Fatalf("failed when converting to struct")
	}
	json.Unmarshal(jsonResult, &choosenTopics)

	tokenSubcriber := createUserAndLogin(t)
	jsonReader = strings.NewReader(fmt.Sprintf(`{
		"topic_id": %d
	}`, choosenTopics.Id))
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/subscribers", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + tokenSubcriber,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(CreateSubscribeHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("Want 200 but return %v", reply.Message)
	}
	return
}
func TestCreateNotificationHandler(t *testing.T) {
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader(`{
		"title": "example topic",
		"desc": "create topic test"
	}`)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	fw := functionWrapper(CreateTopicHandler)
	wrappedFunc := TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}

	jsonReader = strings.NewReader(`{
		"title": "example topic",
		"description": "create topic test for subscriber"
	}`)
	req = httptest.NewRequest(http.MethodGet, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(GetTopicHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("want 200 but get %v", reply.Code)
	}
	if reply.Message == nil {
		t.Fatalf("Message should not be nil")
	}

	parsed := reply.Message.([]interface{})
	choosenTopics := dba.Topic{}
	jsonResult, err := json.Marshal(parsed[0])
	if err != nil {
		t.Fatalf("failed when converting to struct")
	}
	json.Unmarshal(jsonResult, &choosenTopics)

	subscriberToken := createUserAndLogin(t)
	jsonReader = strings.NewReader(fmt.Sprintf(`{
		"topic_id": %d
	}`, choosenTopics.Id))
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/subscribe", jsonReader) 
	req.Header["Authentication"] = []string{"Bearer " + subscriberToken,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(CreateSubscribeHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("want 200 but return %v", reply.Message)
	}

	jsonReader = strings.NewReader(fmt.Sprintf(`{
		"topic_id": %d,
		"message": "test 1"
	}`, choosenTopics.Id))
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/notifications", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	createNotification := CreateNotification{}
	wrappedFunc = TokenCheckMiddleware(createNotification)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("Want 200 but return %v", reply.Message)
	}
	return
}
func TestGetNotificationHandler(t *testing.T) {
	/**
		Test Steps:
		1. Get the created topic
		2. create a subscription
		3. topic creator send notification
		4. subscriber get notification 
	*/
	
	token := createUserAndLogin(t)
	jsonReader := strings.NewReader(`{
		"title": "example topic",
		"desc": "create topic test"
	}`)
	req := httptest.NewRequest(http.MethodPost, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w := httptest.NewRecorder()
	fw := functionWrapper(CreateTopicHandler)
	wrappedFunc := TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	reply := HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}
	// Create a topic
	jsonReader = strings.NewReader("")
	req = httptest.NewRequest(http.MethodGet, baseUrl + "/topics", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(GetTopicHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("want 200 but get %v", reply.Message)
	}
	if reply.Message == nil {
		t.Fatalf("Message should not be nil")
	}

	parsed := reply.Message.([]interface{})
	choosenTopics := dba.Topic{}
	t.Logf("=================%v", parsed)
	jsonResult, err := json.Marshal(parsed[0])
	if err != nil {
		t.Fatalf("failed when converting to struct")
	}
	json.Unmarshal(jsonResult, &choosenTopics)

	// create a subscription
	subscriberToken := createUserAndLogin(t)
	jsonReader = strings.NewReader(fmt.Sprintf(`{
		"topic_id": %d
	}`, choosenTopics.Id))
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/subscribe", jsonReader) 
	req.Header["Authentication"] = []string{"Bearer " + subscriberToken,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(CreateSubscribeHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("want 200 but return %v", reply.Message)
	}

	// create a notification
	jsonReader = strings.NewReader(fmt.Sprintf(`{
		"topic_id": %d,
		"message": "test 1"
	}`, choosenTopics.Id))
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/notifications", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	createNotification := CreateNotification{}
	wrappedFunc = TokenCheckMiddleware(createNotification)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("Want 200 but return %v", reply)
	}
	
	// get the notification from subsriber side
	req = httptest.NewRequest(http.MethodGet, baseUrl + "/notifications", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + subscriberToken,}
	req.Header["Content-Type"] = []string{"application/json",}
	w = httptest.NewRecorder()
	fw = functionWrapper(GetNotificationHandler)
	wrappedFunc = TokenCheckMiddleware(fw)
	wrappedFunc.ServeHTTP(w, req)
	reply = extractReply(w)
	if reply.Code != http.StatusOK {
		t.Fatalf("Want 200 but return %v", reply.Message)
	}
	infer := reply.Message.(string)
	message, err := b64.StdEncoding.DecodeString(infer)
	if err != nil {
		t.Fatalf("Should succeed read base64 reply")
	}
	messageStruct := dba.Notifications{}
	if err := json.Unmarshal(message, &messageStruct); err != nil {
		t.Fatalf("Should be able to read message")
	}
	if (len(messageStruct) == 0) {
		t.Fatalf("Message should have member")
	}

	actualNotification := messageStruct[0].Message 
	if (actualNotification != "test 1") {
		t.Fatalf("Message not equal to test 1")
	}
	
	return
}