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
    err := encoder.Encode(u.Id)
    if err!=nil {
        return nil, err
    }
    err = encoder.Encode(u.Created)
    if err!=nil {
        return nil, err
    }
    err = encoder.Encode(u.Updated)
    if err!=nil {
        return nil, err
    }
    err = encoder.Encode(u.Email)
    if err!=nil {
        return nil, err
    }
    err = encoder.Encode(u.PasswordHash)
    if err!=nil {
        return nil, err
    }
    return w.Bytes(), nil
}

func (u *userInfo) gobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    err := decoder.Decode(&u.Id)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&u.Created)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&u.Updated)
    if err!=nil {
        return err
    }
    err = decoder.Decode(&u.Email)
    if err!=nil {
        return err
    }
    return decoder.Decode(&u.PasswordHash)
}


type newUserPostReq struct {
  Email string `json:"email"`
  Password string `json:"password"`
}

type userGetReq struct {
  Email string `json:"email"`
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

  err=UpdKV([]byte(userPost.Email),userEncoded,[]byte("users"))
  if err!=nil{
    log.Printf("PostNewUser UpdKV %v error",err)
    http.Error(w,"{\"status\" : \"user update error\"}",http.StatusBadRequest)
    return    
  }

  w.WriteHeader(http.StatusCreated);
  w.Write([]byte("{\"status\" : \"ok\"}"))
}


// curl -v -XGET -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com"}' http://localhost:8080/api/v1/users

func GetUser(w http.ResponseWriter, r *http.Request){
  
  content, err := ioutil.ReadAll(r.Body)
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

  err,res:=GetKV([]byte(userGetReq.Email),[]byte("users"))
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

  user := &userInfo{}
  err = user.gobDecode(res)
  if err!=nil{
    log.Printf("GetUser user.gobDecode %v error",err)
    http.Error(w,"{\"status\" : \"user decode error\"}",http.StatusBadRequest)
    return
  }
/*
  user.Status = "ok"
  bin, err := json.MarshalIndent(user, "", "  ")
  if err != nil {
    log.Printf("GetUser json.MarshalIndent error %v",err)
    http.Error(w,"{\"status\" : \"user json error\"}",http.StatusBadRequest)
    return
  }
  */
  var js string
  js=fmt.Sprintf("{\n\"id\":\"%s\",\n\"created\":\"%s\",\n\"email\":\"%s\",\n\"status\":\"ok\"\n}",user.Id,user.Created.Format(time.RFC1123),user.Email)
  w.WriteHeader(http.StatusOK);
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.Write([]byte(js))  
}




