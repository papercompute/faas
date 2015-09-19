package faas

import (
	"bytes"
	"encoding/gob"
//	"encoding/json"
//	"encoding/binary"
	"fmt"
//	"io/ioutil"
	"log"
	"net/http"
	"time"
	"errors"
//	"strings"
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




func gobEncodeI(v interface{})([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func gobDecodeI(buf []byte, v interface{}) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	return decoder.Decode(v)
}

func getIFromDB(key []byte, v interface{}, bucket []byte ) (error){
	err, res := GetKV(key, bucket)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("Error getIFromDB empty key")
	}

	err = gobDecodeI(res, v)
	if err != nil {
		return err
	}

	return nil
}

func setIToDB(key []byte, v interface{}, bucket []byte) (error){

	encoded, err := gobEncodeI(v)
	if err != nil {
		return err
	}

	err = UpdKV(key, encoded, bucket)
	if err != nil {
		return err
	}

	return nil
}

type testInfo struct{
	Name string
	Created time.Time
	Tested bool
	Id string
	Number int64
	Arr []byte
}

func TestIDB() error {

	ti:=&testInfo{
		Name:"Begemot",
		Created:time.Now(),
		Tested:true,
		Id:NewUUID(),
		Number:1235172376152376152,
		Arr:[]byte("Hero never give up"),
	}

	err:=setIToDB([]byte("hero1"),&ti,[]byte(BucketTests))
	if err!=nil{
		return err
	}

	ti2:=&testInfo{}
	err=getIFromDB([]byte("hero1"),&ti2,[]byte(BucketTests))
	if err!=nil{
		return err
	}

	log.Printf("%v",ti)
	log.Printf("%v",ti2)


	if ti.Name != ti2.Name{
		return errors.New(fmt.Sprintf("incorrect Name %s,%s",ti.Name,ti2.Name))	
	}
	if ti.Created != ti2.Created{
		return errors.New(fmt.Sprintf("incorrect Created %v,%v",ti.Created,ti2.Created))	
	}
	if ti.Tested != ti2.Tested{
		return errors.New(fmt.Sprintf("incorrect Tested %v,%v",ti.Tested,ti2.Tested))	
	}
	if ti.Id != ti2.Id{
		return errors.New(fmt.Sprintf("incorrect Id %v,%v",ti.Id,ti2.Id))	
	}
	if ti.Number != ti2.Number{
		return errors.New(fmt.Sprintf("incorrect Number %v,%v",ti.Number,ti2.Number))	
	}
	for i,_:=range ti.Arr{ 
	  if ti.Arr[i] != ti2.Arr[i]{
		return errors.New(fmt.Sprintf("incorrect Arr %d, %v,%v",i,ti.Arr[i],ti2.Arr[i]))	
	  }
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
	err := ReadJSON(r,&userPost)
	if err!=nil{
		log.Printf("RegisterNewUser  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

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
	rec:=confirmUserEmailRec{
		Email:user.Email,
		Created:time.Now(),
	}

	err = setIToDB([]byte(emailConfirmationId),&rec,[]byte(BucketAwaitEmailConfirmationIds))
	if err != nil {
		log.Printf("RegisterNewUser setIToDB %s : %s error %v", user.Email, emailConfirmationId, err)
		http.Error(w, "{\"status\" : \"internal server error\"}", http.StatusInternalServerError)
		return
	}


	go func () {
	  SendMail(user.Email,"registration confirmation link","Follow this link "+CFG.Url+"/api/v1/users/confirm/email/"+emailConfirmationId)
	}()		
	
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("{\n\"id\":\"%s\",\n\"status\":\"ok\"\n}", user.Id)))
}

type userGetReq struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

// curl -v -XGET -H "Content-Type: application/json" 
// -H "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Imtvc21vZGJAZ21haWwuY29tIiwiZXhwIjoxNDQyNDMyNjQxfQ.nqcAjw8C0MKLPyqlmhuMIhS1bN7Z75aOzOUBqomXxRI" 
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
		http.Error(w, "{\"status\" : \"invalid token\"}", http.StatusBadRequest)
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
		http.Error(w, "{\"status\" : \"user email is not confirmed\"}", http.StatusBadRequest)
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
	err := ReadJSON(r,&loginPost)
	if err!=nil{
		log.Printf("LoginUser  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}


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
	
	//log.Printf("user %v",user)

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


// curl -v -XGET -H "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Imtvc21vZGJAZ21haWwuY29tIiwiZXhwIjoxNDQyNDMyNjQxfQ.nqcAjw8C0MKLPyqlmhuMIhS1bN7Z75aOzOUBqomXxRI" 
// http://localhost:8080/api/v1/users/refresh
func RefreshUserToken(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	authToken := r.Header.Get("X-Auth-Token")
	if authToken == "" {
		log.Printf("RefreshUserToken X-Auth-Token error")
		http.Error(w, "{\"status\" : \"token error\"}", http.StatusBadRequest)
		return
	}

	email, err:= CheckAuthToken(authToken)
	if err != nil {
		log.Printf("RefreshUserToken check X-Auth-Token error %v", err)
		http.Error(w, "{\"status\" : \"invalid token\"}", http.StatusBadRequest)
		return
	}

	var user *userInfo
	err,user = getUserInfoFromDB(email)
	if err != nil {
		log.Printf("RefreshUserToken user %s not found", email)
		http.Error(w, "{\"status\" : \"user not found\"}", http.StatusNotFound)
		return
	}

	if user.Confirmed == zeroTime {
		log.Printf("RefreshUserToken email not confirmed")
		http.Error(w, "{\"status\" : \"user email is not confirmed\"}", http.StatusBadRequest)
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




type confirmUserEmailRec struct{
	Email string
	Created time.Time
}


// curl -v -XPOST http://localhost:8080/api/v1/users/confirm/:id
func ConfirmUserEmail(w http.ResponseWriter, r *http.Request, id string) {

	var err error

	w.Header().Set("Content-Type", "application/json")

	bucket:=[]byte(BucketAwaitEmailConfirmationIds)


	rec:=confirmUserEmailRec{}

	err = getIFromDB([]byte(id),&rec,bucket)
	if err != nil {
		log.Printf("ConfirmUserEmail getIFromDB %s error %v", id, err)
		http.Error(w, "{\"status\" : \"internal server error\"}", http.StatusBadRequest)
		return
	}

	email:=rec.Email

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


// curl -v -XPOST -H "Content-Type: application/json" 
// -d '{"email":"sobaka@drug.com","password":"123456789"}' 
// http://localhost:8080/api/v1/users/resend/confirm/email
func ResendUserConfirmationEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	loginPost := loginPostReq{}
	err := ReadJSON(r,&loginPost)
	if err!=nil{
		log.Printf("ResendConfirmUserEmail  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}
	var user *userInfo
	err,user = getUserInfoFromDB(loginPost.Email)
	if err != nil {
		log.Printf("ResendConfirmUserEmail user %s not found", loginPost.Email)
		http.Error(w, "{\"status\" : \"not found\"}", http.StatusNotFound)
		return
	}

	if err=CompareBcryptHashAndPassword(user.PasswordHash, loginPost.Password); err!=nil{
		log.Printf("ResendConfirmUserEmail wrong password")
		http.Error(w, "{\"status\" : \"wrong password\"}", http.StatusBadRequest)
		return
	}

	if user.Confirmed != zeroTime {
		log.Printf("ResendConfirmUserEmail user email has already confirmed %v",user.Confirmed)
		http.Error(w, "{\"status\" : \"user email has already confirmed\"}", http.StatusBadRequest)
		return		
	}
	
	emailConfirmationId:=NewUUID()
	rec:=confirmUserEmailRec{
		Email:user.Email,
		Created:time.Now(),
	}

	err = setIToDB([]byte(emailConfirmationId),&rec,[]byte(BucketAwaitEmailConfirmationIds))
	if err != nil {
		log.Printf("ResendConfirmUserEmail setIToDB %s : %s error %v", user.Email, emailConfirmationId, err)
		http.Error(w, "{\"status\" : \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	go func () {
	  SendMail(user.Email,"registration confirmation link",
	  	"Follow this link "+CFG.Url+"/api/v1/users/confirm/email/"+emailConfirmationId)
	}()		

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))	
}




type passResetGetReq struct {
	Email    string `json:"email"`
}


type passResetTokenRec struct{
	Email string
	Created time.Time
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com"}' 
// http://localhost:8080/api/v1/users/password/sendResetToken
func SendPasswordResetTokenToUserEmail(w http.ResponseWriter, r *http.Request) {	
	w.Header().Set("Content-Type", "application/json")

	var err error

	passResetGet := passResetGetReq{}

	if err = ReadJSON(r,&passResetGet); err!=nil{
		log.Printf("SendPasswordResetTokenToUserEmail  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

	passwordResetToken:=NewUUID()

	rec:=passResetTokenRec{
		Email:passResetGet.Email,
		Created:time.Now(),
	}

	err = setIToDB([]byte(passwordResetToken),&rec,[]byte(BucketPasswordResetIds))
	if err != nil {
		log.Printf("SendPasswordResetTokenToUserEmail setIToDB %s error %v", passwordResetToken, err)
		http.Error(w, "{\"status\" : \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	go func () {
	  SendMail(passResetGet.Email,"password reset token",
	  	"Copy this token to password reset form "+passwordResetToken)
	}()		


	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}


type passResetByTokenPostReq struct {
	Email    string `json:"email"`
	Password    string `json:"password"`
	Token    string `json:"token"`
}


// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789","token":"12312312312312"}' 
// http://localhost:8080/api/v1/users/password/reset
func ResetUserPasswordWithNewOneByToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	passPost := passResetByTokenPostReq{}
	err := ReadJSON(r,&passPost)
	if err!=nil{
		log.Printf("ResetUserPasswordWithNewOneByToken  ReadJSON error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

	if len(passPost.Email) < 4 { // a@a.co
		log.Printf("ResetUserPasswordWithNewOneByToken email error")
		http.Error(w, "{\"status\" : \"email too short\"}", http.StatusBadRequest)
		return
	}
	if len(passPost.Password) < 4 { // a@a.co
		log.Printf("ResetUserPasswordWithNewOneByToken password error")
		http.Error(w, "{\"status\" : \"password too short\"}", http.StatusBadRequest)
		return
	}

	rec:=passResetTokenRec{}

	err = getIFromDB([]byte(passPost.Token),&rec,[]byte(BucketPasswordResetIds))
	if err != nil {
		log.Printf("ResetUserPasswordWithNewOneByToken getIFromDB %s error %v", passPost.Token, err)
		http.Error(w, "{\"status\" : \"internal server error\"}", http.StatusInternalServerError)
		return
	}

	if rec.Email!=passPost.Email{
		log.Printf("ResetUserPasswordWithNewOneByToken wrong email %s error %s : %s", passPost.Token, rec.Email, passPost.Email)
		http.Error(w, "{\"status\" : \"wrong email\"}", http.StatusBadRequest)
		return
	}

	var user *userInfo
	err,user = getUserInfoFromDB(passPost.Email)
	if err != nil {
		log.Printf("ResetUserPasswordWithNewOneByToken user %s not found", passPost.Email)
		http.Error(w, "{\"status\" : \"not found\"}", http.StatusNotFound)
		return
	}


	if user.Confirmed == zeroTime {
		log.Printf("ResetUserPasswordWithNewOneByToken email not confirmed")
		http.Error(w, "{\"status\" : \"user not confirmed email\"}", http.StatusBadRequest)
		return		
	}	

	user.PasswordHash = GetBcryptHash(passPost.Password)

	err = setUserInfoToDB(user)
	if err != nil {
		log.Printf("ResetUserPasswordWithNewOneByToken setUserInfoToDB error %v", err)
		http.Error(w, "{\"status\" : \"set user error\"}", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}

type changePassPostReq struct {
	Email    string `json:"email"`
	OldPassword string `json:"oldpassword"`
	NewPassword string `json:"newpassword"`
}


// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","oldpassword":"123456789","newpassword":"123456789"}'
// http://localhost:8080/api/v1/users/password/change 
func ChangeUserPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	changePassPost := changePassPostReq{}

	if err = ReadJSON(r,&changePassPost); err!=nil{
		log.Printf("ChangeUserPassword  json error %v", err)
		http.Error(w, "{\"status\" : \"bad reguest\"}", http.StatusBadRequest)
		return
	}

	if len(changePassPost.NewPassword) < 4 { // a@a.co
		log.Printf("ChangeUserPassword password error")
		http.Error(w, "{\"status\" : \"password too short\"}", http.StatusBadRequest)
		return
	}


	var user *userInfo
	err,user = getUserInfoFromDB(changePassPost.Email)
	if err != nil {
		log.Printf("ChangeUserPassword user %s not found", changePassPost.Email)
		http.Error(w, "{\"status\" : \"not found\"}", http.StatusNotFound)
		return
	}


	if user.Confirmed == zeroTime {
		log.Printf("ChangeUserPassword email not confirmed")
		http.Error(w, "{\"status\" : \"user not confirmed email\"}", http.StatusBadRequest)
		return		
	}	

	if err=CompareBcryptHashAndPassword(user.PasswordHash, changePassPost.OldPassword); err!=nil{
		log.Printf("ChangeUserPassword wrong password")
		http.Error(w, "{\"status\" : \"wrong password\"}", http.StatusBadRequest)
		return
	}

	user.PasswordHash = GetBcryptHash(changePassPost.NewPassword)

	err = setUserInfoToDB(user)
	if err != nil {
		log.Printf("RegisterNewUser setUserInfoToDB error %v", err)
		http.Error(w, "{\"status\" : \"set user error\"}", http.StatusInternalServerError)
		return
	}


	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}

