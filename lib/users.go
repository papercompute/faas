package faas

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
//	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"errors"
)

var (

	zeroTime = time.Time{}
)

type userInfo struct {
	Id           string    `json:"id"`
	Created      time.Time `json:"created"`
	Updated      time.Time `json:"updated"`
	Email        string    `json:"email"`
	Confirmed    time.Time `json:"confirmed"`
	PasswordHash string    `json:"-"`
}




func (u *userInfo) gobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(*u)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (u *userInfo) gobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(u)
}


func getUserInfoFromDB(email string) (error, *userInfo){
	err, res := GetKV([]byte(email), []byte(BucketUsers))
	if err != nil {
		return err, nil
	}
	if res == nil {
		return errors.New("Error no user info"),nil
	}

	user := &userInfo{}
	err = user.gobDecode(res)
	if err != nil {
		return err,nil
	}

	return nil,user
}

func setUserInfoToDB(u *userInfo) (error){

	userEncoded, err := u.gobEncode()
	if err != nil {
		return err
	}

	err = UpdKV([]byte(u.Email), userEncoded, []byte(BucketUsers))
	if err != nil {
		return err
	}

	return nil
}


type newUserPostReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"hero@dream.com","password":"123456789"}' http://localhost:8080/api/v1/users/register
func RegisterNewUser(w http.ResponseWriter, r *http.Request) {


	w.Header().Set("Content-Type", "application/json")

	userPost := newUserPostReq{}

	if err := ReadJSON(r,userPost); err!=nil{
		log.Printf("RegisterNewUser  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

/*	
	var content []byte
	if content, err = ioutil.ReadAll(r.Body); err != nil {
		log.Printf("RegisterNewUser ioutil.ReadAll error: %v", err)
		http.Error(w, "{\"status\" : \"request data error\"}", http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(content, &userPost); err != nil {
		log.Printf("RegisterNewUser json.Unmarshal error: %v %s", err, string(content))
		http.Error(w, "{\"status\" : \"request json error\"}", http.StatusBadRequest)
		return
	}
*/	

	if len(userPost.Email) < 4 { // a@a.co
		log.Printf("RegisterNewUser email error")
		http.Error(w, "{\"status\" : \"email too short\"}", http.StatusBadRequest)
		return
	}
	if len(userPost.Password) < 4 { // a@a.co
		log.Printf("RegisterNewUser password error")
		http.Error(w, "{\"status\" : \"password too short\"}", http.StatusBadRequest)
		return
	}

	// check email existence

	var res []byte
	err, res = GetKV([]byte(userPost.Email), []byte(BucketUsers))
	if err != nil {
		log.Printf("RegisterNewUser GetKV error %v", err)
		http.Error(w, "{\"status\" : \"key "+userPost.Email+" error\"}", http.StatusBadRequest)
		return
	}
	if res != nil {
		log.Printf("RegisterNewUser email %s already exists", userPost.Email)
		http.Error(w, "{\"status\" : \"email "+userPost.Email+" already exists\"}", http.StatusBadRequest)
		return
	}

	user := &userInfo{
		Created:      time.Now(),
		Updated:      time.Now(),
		Email:        userPost.Email,
		Confirmed:    zeroTime,
		PasswordHash: GetBcryptHash(userPost.Password),
	}

	err = setUserInfoToDB(user)
	if err != nil {
		log.Printf("RegisterNewUser setUserInfoToDB error %v", err)
		http.Error(w, "{\"status\" : \"set user error\"}", http.StatusInternalServerError)
		return
	}

	emailConfirmationId:=NewUUID()

	err = UpdKV([]byte(emailConfirmationId), []byte(user.Email+":"+string(time.Now().Unix())), []byte(BucketAwaitEmailConfirmationIds))
	if err != nil {
		log.Printf("RegisterNewUser UpdKV %s : %s error %v", user.Email, emailConfirmationId, err)
		http.Error(w, "{\"status\" : \"email confirmation id management error\"}", http.StatusInternalServerError)
		return
	}

	go func () {
	  SendMail(user.Email,"registration confirmation link","Follow this link "+CFG.Url+"/api/v1/users/confirm/email/"+emailConfirmationId)
	}()		
	
/*
	err = SendMail(user.Email,"registration confirmation link","Follow this link "+CFG.Url+"/api/v1/users/confirm/"+emailConfirmationId)
	if err != nil {
		log.Printf("RegisterNewUser sendmail %s error %v", user.Email, err)
		http.Error(w, "{\"status\" : \"sendmail error\"}", http.StatusInternalServerError)
		return
	}
*/
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("{\n\"id\":\"%s\",\n\"status\":\"ok\"\n}", user.Id)))
}

type userGetReq struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

// curl -v -XGET -H "Content-Type: application/json" -H "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Imtvc21vZGJAZ21haWwuY29tIiwiZXhwIjoxNDQyNDMyNjQxfQ.nqcAjw8C0MKLPyqlmhuMIhS1bN7Z75aOzOUBqomXxRI" 
// http://localhost:8080/api/v1/users/info
func GetUserInfo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	authToken := r.Header.Get("X-Auth-Token")
	if authToken == "" {
		log.Printf("GetUser X-Auth-Token error")
		http.Error(w, "{\"status\" : \"token error\"}", http.StatusBadRequest)
		return
	}

	email, err:= CheckAuthToken(authToken)
	if err != nil {
		log.Printf("GetUser check X-Auth-Token error %v", err)
		http.Error(w, "{\"status\" : \"check token error\"}", http.StatusBadRequest)
		return
	}

	var user *userInfo
	err,user = getUserInfoFromDB(email)
	if err != nil {
		log.Printf("GetUser user %s not found", email)
		http.Error(w, "{\"status\" : \"user not found\"}", http.StatusNotFound)
		return
	}

	if user.Confirmed == zeroTime {
		log.Printf("GetUser email not confirmed")
		http.Error(w, "{\"status\" : \"user not confirmed email\"}", http.StatusBadRequest)
		return		
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("{\n\"created\":\"%s\",\n\"email\":\"%s\",\n\"status\":\"ok\"\n}",
		user.Created.Format(time.RFC1123), user.Email)))
}

type loginPostReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users/login
func LoginUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	loginPost := loginPostReq{}

	if err := ReadJSON(r,loginPost); err!=nil{
		log.Printf("LoginUser  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

/*	
	var content []byte
	if content, err = ioutil.ReadAll(r.Body); err != nil {
		log.Printf("LoginUser ioutil.ReadAll error: %v", err)
		http.Error(w, "{\"status\" : \"request data error\"}", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(content, &loginPost); err != nil {
		log.Printf("LoginUser json.Unmarshal error: %v %s", err, string(content))
		http.Error(w, "{\"status\" : \"request json error\"}", http.StatusBadRequest)
		return
	}
*/	

	if len(loginPost.Email) < 4 { // a@a.co
		log.Printf("LoginUser email error")
		http.Error(w, "{\"status\" : \"email too short\"}", http.StatusBadRequest)
		return
	}
	if len(loginPost.Password) < 4 { // a@a.co
		log.Printf("LoginUser password error")
		http.Error(w, "{\"status\" : \"password too short\"}", http.StatusBadRequest)
		return
	}
	var user *userInfo
	err,user = getUserInfoFromDB(loginPost.Email)
	if err != nil {
		log.Printf("LoginUser user %s not found", loginPost.Email)
		http.Error(w, "{\"status\" : \"not found\"}", http.StatusNotFound)
		return
	}


	if user.Confirmed == zeroTime {
		log.Printf("LoginUser email not confirmed")
		http.Error(w, "{\"status\" : \"user not confirmed email\"}", http.StatusBadRequest)
		return		
	}
	
	log.Printf("user %v",user)

	if err=CompareBcryptHashAndPassword(user.PasswordHash, loginPost.Password); err!=nil{
		log.Printf("LoginUser wrong password")
		http.Error(w, "{\"status\" : \"wrong password\"}", http.StatusBadRequest)
		return
	}

	token, err := GetAuthToken(user)
	if err != nil {
		log.Printf("LoginUser GetAuthToken error %v", err)
		http.Error(w, "{\"status\" : \"user decode error\"}", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\n\"token\":\"%s\",\n\"status\":\"ok\"\n}", token)))
}



// curl -v -XPOST http://localhost:8080/api/v1/users/confirm/:userid
func ConfirmUserEmail(w http.ResponseWriter, r *http.Request, id string) {

	w.Header().Set("Content-Type", "application/json")

	bucket:=[]byte(BucketAwaitEmailConfirmationIds)
	
	err, res := GetKV([]byte(id), bucket)
	if err != nil {
		log.Printf("ConfirmUserEmail GetKV error %v", err)
		http.Error(w, "{\"status\" : \"get id error\"}", http.StatusBadRequest)
		return
	}
	if res == nil {
		log.Printf("ConfirmUserEmail confirm email id %s not found", id)
		http.Error(w, "{\"status\" : \"id not found\"}", http.StatusNotFound)
		return
	}

	d:=strings.Split(string(res),":")
	email:=d[0]
/*
	if t0, err := strconv.ParseInt(d[1], 10, 64); err != nil {
		log.Printf("ConfirmUserEmail bad data")
		http.Error(w, "{\"status\" : \"bad data\"}", http.StatusBadRequest)
		return
	}	
*/
	var user *userInfo
	err,user = getUserInfoFromDB(email)
	if err != nil {
		log.Printf("ConfirmUserEmail user %s not found", email)
		http.Error(w, "{\"status\" : \"not found\"}", http.StatusNotFound)
		return
	}

    if err = DelKV([]byte(id), bucket); err != nil {
		log.Printf("ConfirmUserEmail DelKV error %v", err)
		http.Error(w, "{\"status\" : \"del id error\"}", http.StatusInternalServerError)
		return
	}

	user.Confirmed = time.Now()

	err = setUserInfoToDB(user)
	if err != nil {
		log.Printf("ConfirmUserEmail setUserInfoToDB error %v", err)
		http.Error(w, "{\"status\" : \"set user error\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}


func ResendConfirmUserEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))	
}




type passResetPostReq struct {
	Email    string `json:"email"`
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com"}' 
// http://localhost:8080/api/v1/users/password/sendResetToken
func SendPasswordResetTokenToUserEmail(w http.ResponseWriter, r *http.Request) {	
	w.Header().Set("Content-Type", "application/json")
	var err error
	passResetPost := passResetPostReq{}

	if err = ReadJSON(r,passResetPost); err!=nil{
		log.Printf("SendPasswordResetTokenToUserEmail  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' 
// http://localhost:8080/api/v1/users/password/reset/3bea3a7ba0814591852016fdc8c3ecce
func ResetUserPasswordWithNewOneByToken(w http.ResponseWriter, r *http.Request, token string) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}

type changePassPostReq struct {
	Email    string `json:"email"`
	OldPassword string `json:"oldpassword"`
	NewPassword string `json:"newpassword"`
}

func ChangeUserPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}
