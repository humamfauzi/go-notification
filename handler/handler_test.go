package handler

import (
	"testing"
	"strings"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	dba "github.com/humamfauzi/go-notification/database"
)

const (
	baseUrl = "http://example.com"
)

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
	t.Log(reply)
	token := reply.Message.(map[string]interface{})["token"].(string)
	jsonReader = strings.NewReader("")
	req = httptest.NewRequest(http.MethodPost, baseUrl + "/users/check", jsonReader)
	req.Header["Authentication"] = []string{"Bearer " + token,}
	req.Header["Content-Type"] = []string{"application/json"}
	w = httptest.NewRecorder()
	checkLogin := CheckLogin{}
	wrappedFunc := TokenCheckMiddleware(checkLogin)
	wrappedFunc.ServeHTTP(w, req)
	// checkLogin.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	reply = HandlerReply{}
	json.Unmarshal(body, &reply)

	if reply.Code != http.StatusOK {
		t.Fatalf("Should return ok but return %v", reply)
	}

	

}

func TestUpdateUserHandler(t *testing.T) {
	return
}
func TestDeleteUserHandler(t *testing.T) {
	return
}
func TestCreateTopicHandler(t *testing.T) {
	return
}
func TestGetTopicHandler(t *testing.T) {
	return
}
func TestCreateSubscribeHandler(t *testing.T) {
	return
}
func TestCreateNotificationHandler(t *testing.T) {
	return
}
func TestGetNotificationHandler(t *testing.T) {
	return
}