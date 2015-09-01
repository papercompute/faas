package main

import (
	"net/http"
	"log"
	"runtime"
"crypto/md5"
    "encoding/hex"
//	"io"
  "io/ioutil"
  "encoding/json"
//	"regexp"
)


func GetMD5Hash(text string) string {
    hasher := md5.New()
    hasher.Write([]byte(text))
    return hex.EncodeToString(hasher.Sum(nil))
}

// http.Error(rw, err.Error(), 500)

func WriteStatus(w http.ResponseWriter, code int, status string) {
  w.WriteHeader(code)
  WriteJSON(w, map[string]string{
    "status": status,
  })
}
func WriteJSON(w http.ResponseWriter, v interface{}) error {
  b, err := json.MarshalIndent(v, "", "  ")
  if err != nil {
    log.Printf("WriteJSON error %s",err)
    return err
  }
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
  w.Write(b)
  return nil
}

func ReadJSON(r *http.Request, v interface{}) error {
  content, err := ioutil.ReadAll(r.Body);
  if  err != nil {
    return err
  }
   
  if err := json.Unmarshal(content, v); err != nil {
    return err
  }

  return nil  
}

type Faas struct {}

type User struct {
  Id uint
  Email string
  PasswordHash string
}

type NewUserPost struct {
  Email string `json:"email"`
  Password string `json:"password"`
}

type UserGet struct {
  Email string `json:"email"`
}

var userIdCounter uint = 1;
var userEmailStore = map[string]*User{}
var userIdStore = map[uint]*User{}

func saveStoresToFile() error{
  b, err := json.MarshalIndent(userEmailStore, "", "  ")
  if err != nil {
    log.Printf("saveStoresToFile error %s",err)
    return err
  }
  ioutil.WriteFile("users",b,0644); 
  return nil
}


func postNewUser(w http.ResponseWriter, r *http.Request){
  userPost:=NewUserPost{}

  if err := ReadJSON(r, &userPost); err != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api newuser post data error")
    return
  }

  log.Printf("userPost{%s,%s}",userPost.Email,userPost.Password)
  if len(userPost.Email) < 4 { // a@a.co
    WriteStatus(w,500,"api newuser email error")
    return
  }
  if len(userPost.Password) < 4 { // a@a.co
    WriteStatus(w,500,"api newuser password error")
    return
  }

  if u:=userEmailStore[userPost.Email]; u!= nil{
    WriteStatus(w,500,"user already exists")
    return
  }

  passwordHash:=GetMD5Hash(userPost.Password)
  user := &User{
    Id:userIdCounter,
    Email:userPost.Email,
    PasswordHash:passwordHash,
  }

  userIdCounter=userIdCounter+1;

  userEmailStore[user.Email]=user
  userIdStore[user.Id]=user

  go saveStoresToFile()

  WriteStatus(w,200,"ok")
}

func getUser(w http.ResponseWriter, r *http.Request){
  content, err := ioutil.ReadAll(r.Body)
  if err != nil {
    WriteStatus(w,500,"api user get error")
    return
  }
  userGet:=UserGet{}
  err = json.Unmarshal(content, &userGet) 
  if err != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api user get data error")
    return
  }

  log.Printf("userGet{%s}",userGet.Email)
  if u:=userEmailStore[userGet.Email]; u!= nil{
    w.WriteHeader(200)
    WriteJSON(w,u)
    return
  }

  WriteStatus(w,500,"api user not found")

}

var fileServer = http.FileServer(http.Dir("public"))

func (e *Faas) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
  switch r.URL.Path {
    case "/api/v1/newuser":
     if r.Method == "POST" {
       postNewUser(w,r)
       return
     }
    case "/api/v1/user":
     if r.Method == "GET" {
       getUser(w,r)
       return
     }

    default:
     if r.Method == "GET" {
       fileServer.ServeHTTP(w,r)
       return
     }
  }
  http.NotFound(w, r)
}




func main() {
  runtime.GOMAXPROCS(runtime.NumCPU() - 1)
  http.ListenAndServe(":8080", &Faas{})
}
