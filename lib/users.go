package faas

import (
  "log"
  "net/http"
  "bytes"
  "encoding/gob"
  "time"  
  "io/ioutil"
  "encoding/json"
  "fmt"
)


type userInfo struct {
  Id *UUID `json:"id"`
  Created time.Time  `json:"created"`
  Updated time.Time  `json:"updated"`
  Email string  `json:"email"`
  PasswordHash string `json:"-"`
}


func (u *userInfo) gobEncode() ([]byte, error) {
    w := new(bytes.Buffer)
    encoder := gob.NewEncoder(w)
    err := encoder.Encode(*u)
    if err!=nil {
        return nil, err
    }
    return w.Bytes(), nil
}

func (u *userInfo) gobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    return decoder.Decode(u)
}


type newUserPostReq struct {
  Email string `json:"email"`
  Password string `json:"password"`
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users
func PostNewUser(w http.ResponseWriter, r *http.Request){
  
  userPost:=newUserPostReq{}

  var err error
  var content []byte  
  if  content, err = ioutil.ReadAll(r.Body); err != nil {
    log.Printf("PostNewUser ioutil.ReadAll error: %v",err)
    http.Error(w,"{\"status\" : \"request data error\"}",http.StatusBadRequest)
    return
  }

  if err = json.Unmarshal(content, &userPost); err != nil {
    log.Printf("PostNewUser json.Unmarshal error: %v %s",err, string(content))
    http.Error(w,"{\"status\" : \"request json error\"}",http.StatusBadRequest)
    return
  }

  if len(userPost.Email) < 4 { // a@a.co
    log.Printf("PostNewUser email error")
    http.Error(w,"{\"status\" : \"email too short\"}",http.StatusBadRequest)
    return
  }
  if len(userPost.Password) < 4 { // a@a.co
    log.Printf("PostNewUser password error")
    http.Error(w,"{\"status\" : \"password too short\"}",http.StatusBadRequest)
    return
  }

  var res []byte
  err,res=GetKV([]byte(userPost.Email),[]byte("users"))
  if err!=nil{
    log.Printf("PostNewUser GetKV error %v",err)
    http.Error(w,"{\"status\" : \"get key "+userPost.Email+" error\"}",http.StatusBadRequest)
    return    
  }
  if res != nil {
    log.Printf("PostNewUser email %s already exists",userPost.Email)
    http.Error(w,"{\"status\" : \"email "+userPost.Email+" already exists\"}",http.StatusBadRequest)
    return
  }

  user := &userInfo{
    Id:NewUUID(),
    Created: time.Now(),//.Format(time.RFC1123),
    Email:userPost.Email,
    PasswordHash:GetMD5Hash(userPost.Password),
  }

  var userEncoded []byte
  userEncoded, err = user.gobEncode()
  if err != nil {
    log.Printf("PostNewUser user.gobEncode %v error",err)
    http.Error(w,"{\"status\" : \"user encode error\"}",http.StatusBadRequest)
    return    
  }

  err=UpdKV([]byte(user.Email),userEncoded,[]byte("users"))
  if err!=nil{
    log.Printf("PostNewUser UpdKV user.Email %v error",err)
    http.Error(w,"{\"status\" : \"user.Email update error\"}",http.StatusBadRequest)
    return    
  }

  err=UpdKV([]byte(user.Id.String()),[]byte(user.Email),[]byte("usersids"))
  if err!=nil{
    log.Printf("PostNewUser UpdKV user.Id %v error",err)
    http.Error(w,"{\"status\" : \"user.Id update error\"}",http.StatusBadRequest)
    return    
  }

  w.WriteHeader(http.StatusCreated);
  w.Write([]byte(fmt.Sprintf("{\n\"id\":\"%s\",\n\"status\":\"ok\"\n}",user.Id)))
}


type userGetReq struct {
  Id string `json:"id"`  
  Email string `json:"email"`
}


// curl -v -XGET -H "Content-Type: application/json" -H "X-Auth-Token: ..." http://localhost:8080/api/v1/users
func GetUser(w http.ResponseWriter, r *http.Request){

  var id string
  var err error
  authToken:=r.Header.Get("X-Auth-Token")
  if authToken==""{
    log.Printf("GetUser X-Auth-Token error")
    http.Error(w,"{\"status\" : \"token error\"}",http.StatusBadRequest)
    return    
  }

  log.Printf("X-Auth-Token %s",authToken)

  var email string
  id,email,err=CheckAuthToken(authToken)
  if err != nil || id=="" {
    log.Printf("GetUser check X-Auth-Token error %v",err)
    http.Error(w,"{\"status\" : \"check token error\"}",http.StatusBadRequest)
    return    
  }

/*
  var content []byte
  content, err = ioutil.ReadAll(r.Body)
  if err != nil {
    log.Printf("GetUser ioutil.ReadAll error %v",err)
    http.Error(w,"{\"status\" : \"read data error\"}",http.StatusBadRequest)
    return    
  }

  userGetReq:=userGetReq{}
  err = json.Unmarshal(content, &userGetReq) 
  if err != nil {
    log.Printf("GetUser json.Unmarshal %v error",err)
    http.Error(w,"{\"status\" : \"json error\"}",http.StatusBadRequest)
    return
  }

  var res []byte

  if userGetReq.Id!=""{

   err,res=GetKV([]byte(userGetReq.Id),[]byte("usersids"))
   if err!=nil{
     log.Printf("GetUser GetKV %s error",userGetReq.Id)
     http.Error(w,"{\"status\" : \"get "+userGetReq.Id+" error\"}",http.StatusBadRequest)
     return    
   }  
   if res == nil {
     log.Printf("GetUser GetKV %s nil",userGetReq.Email)
     http.Error(w,"{\"status\" : \"user not found\"}",http.StatusNotFound)
     return
   }

   userGetReq.Email=string(res)  

  }
  
  if userGetReq.Email!=""{ 
   err,res=GetKV([]byte(userGetReq.Email),[]byte("users"))
   if err!=nil{
    log.Printf("GetUser GetKV %s error",userGetReq.Email)
    http.Error(w,"{\"status\" : \"get "+userGetReq.Email+" error\"}",http.StatusBadRequest)
    return    
   }  
   if res == nil {
    log.Printf("GetUser GetKV %s nil",userGetReq.Email)
    http.Error(w,"{\"status\" : \"user not found\"}",http.StatusNotFound)
    return
   }

  } else {
    log.Printf("GetUser bad request params")
    http.Error(w,"{\"status\" : \"bad request\"}",http.StatusBadRequest)
    return    
  }
  */

  var res []byte

   err,res=GetKV([]byte(email),[]byte("users"))
   if err!=nil{
    log.Printf("GetUser GetKV %s error",email)
    http.Error(w,"{\"status\" : \"get "+email+" error\"}",http.StatusBadRequest)
    return    
   }  
   if res == nil {
    log.Printf("GetUser GetKV %s nil",email)
    http.Error(w,"{\"status\" : \"user not found\"}",http.StatusNotFound)
    return
   }


  user := &userInfo{}
  err = user.gobDecode(res)
  if err!=nil{
    log.Printf("GetUser user.gobDecode %v error",err)
    http.Error(w,"{\"status\" : \"user decode error\"}",http.StatusBadRequest)
    return
  }
  w.WriteHeader(http.StatusOK);
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.Write([]byte(fmt.Sprintf("{\n\"id\":\"%s\",\n\"created\":\"%s\",\n\"email\":\"%s\",\n\"status\":\"ok\"\n}",
    user.Id,user.Created.Format(time.RFC1123),user.Email)))  
}


type loginPostReq struct {
  Email string `json:"email"`
  Password string `json:"password"`
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users/login
func LoginUser(w http.ResponseWriter, r *http.Request){
  
  loginPost:=loginPostReq{}

  var err error
  var content []byte  
  if  content, err = ioutil.ReadAll(r.Body); err != nil {
    log.Printf("LoginUser ioutil.ReadAll error: %v",err)
    http.Error(w,"{\"status\" : \"request data error\"}",http.StatusBadRequest)
    return
  }

  if err = json.Unmarshal(content, &loginPost); err != nil {
    log.Printf("LoginUser json.Unmarshal error: %v %s",err, string(content))
    http.Error(w,"{\"status\" : \"request json error\"}",http.StatusBadRequest)
    return
  }

  if len(loginPost.Email) < 4 { // a@a.co
    log.Printf("LoginUser email error")
    http.Error(w,"{\"status\" : \"email too short\"}",http.StatusBadRequest)
    return
  }
  if len(loginPost.Password) < 4 { // a@a.co
    log.Printf("LoginUser password error")
    http.Error(w,"{\"status\" : \"password too short\"}",http.StatusBadRequest)
    return
  }

  var res []byte
  err,res=GetKV([]byte(loginPost.Email),[]byte("users"))
  if err!=nil{
    log.Printf("LoginUser GetKV error %v",err)
    http.Error(w,"{\"status\" : \"get key "+loginPost.Email+" error\"}",http.StatusBadRequest)
    return    
  }

  user := &userInfo{}
  err = user.gobDecode(res)
  if err!=nil{
    log.Printf("LoginUser user.gobDecode %v error",err)
    http.Error(w,"{\"status\" : \"user decode error\"}",http.StatusBadRequest)
    return
  }

  // now user

  PasswordHash:=GetMD5Hash(loginPost.Password)
  if user.PasswordHash!=PasswordHash{
    log.Printf("LoginUser wrong password")
    http.Error(w,"{\"status\" : \"wrong password\"}",http.StatusBadRequest)
    return
  }

  token,err:=GetAuthToken(user)
  if err!=nil{
    log.Printf("LoginUser GetAuthToken error %v",err)
    http.Error(w,"{\"status\" : \"user decode error\"}",http.StatusBadRequest)
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write([]byte(fmt.Sprintf("{\n\"token\":\"%s\",\n\"status\":\"ok\"\n}",token)))
}


type logoutPostReq struct {
  Token string `json:"token"`
}

// curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users/login
func LogoutUser(w http.ResponseWriter, r *http.Request){
  
  logoutPost:=logoutPostReq{}

  var err error
  var content []byte  
  if  content, err = ioutil.ReadAll(r.Body); err != nil {
    log.Printf("LogoutUser ioutil.ReadAll error: %v",err)
    http.Error(w,"{\"status\" : \"request data error\"}",http.StatusBadRequest)
    return
  }

  if err = json.Unmarshal(content, &logoutPost); err != nil {
    log.Printf("LogoutUser json.Unmarshal error: %v %s",err, string(content))
    http.Error(w,"{\"status\" : \"request json error\"}",http.StatusBadRequest)
    return
  }

}

