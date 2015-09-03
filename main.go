package main

import (
	"net/http"
	"log"
	"runtime"
  "fmt"
"crypto/md5"
    "encoding/hex"
//	"io"
  "io/ioutil"
  "encoding/json"
  "encoding/gob"
  "time"
//	"regexp"
  "github.com/boltdb/bolt"
  "bytes"
//  "runtime"
)

var db *bolt.DB

func UpdKV(key []byte, value []byte, bucket []byte) error {
  return db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket(bucket)
    return b.Put(key, value)
  })
}
func GetKV(key []byte, bucket []byte) (error, []byte) {
  var result []byte = nil
  err := db.View(func(tx *bolt.Tx) error {
    if b := tx.Bucket(bucket); b != nil {
      result = b.Get(key)
    }  
    return nil
  })
  return err, result
}


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
  Id uint `json:"id"`
  Created time.Time  `json:"created"` // time.Now().Unix()
  Email string  `json:"email"`
  PasswordHash string `json:"-"`
}

/*
 var decoded map[string] interface{}
        err := json.Unmarshal(v, &decoded)
        if err != nil {
            return err
        }
        fmt.Printf("Title: %s\n", decoded["Title"])
*/        

func (u *User) GobEncode() ([]byte, error) {
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

func (u *User) GobDecode(buf []byte) error {
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
    err = decoder.Decode(&u.Email)
    if err!=nil {
        return err
    }
    return decoder.Decode(&u.PasswordHash)
}




type NewUserPost struct {
  Email string `json:"email"`
  Password string `json:"password"`
}
type UserGetReq struct {
  Email string `json:"email"`
}
type UserGetResp struct {
  Id uint `json:"id"`
  Created time.Time `json:"created"`
  Email string `json:"email"`
}

var userIdCounter uint = 127 

func GetUserId() uint{
  uid := userIdCounter
  userIdCounter = userIdCounter + 1
  return uid  
}

func postNewUser(w http.ResponseWriter, r *http.Request){
  userPost:=NewUserPost{}

  if err := ReadJSON(r, &userPost); err != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api newuser post data error")
    return
  }

  if len(userPost.Email) < 4 { // a@a.co
    WriteStatus(w,500,"api newuser email error")
    return
  }
  if len(userPost.Password) < 4 { // a@a.co
    WriteStatus(w,500,"api newuser password error")
    return
  }

  err1,res:=GetKV([]byte(userPost.Email),[]byte("users"))
  if res != nil {
    log.Printf("err: %s",err1)
    WriteStatus(w,500,"api newuser user already exists")
    return
  }

//  runtime.Goshced()
  
  user := &User{
    Id:GetUserId(),
    Created: time.Now(), //int32(time.Now().Unix()),
    Email:userPost.Email,
    PasswordHash:GetMD5Hash(userPost.Password),
  }


  encoded, err := user.GobEncode()
  if err != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api newuser user error")
    return
  }


  err0:=UpdKV([]byte(userPost.Email),encoded,[]byte("users"))
  if err0!=nil{
    log.Printf("err: %s",err0)
    WriteStatus(w,500,"api newuser addkv error")
    return
  }


  WriteStatus(w,200,"ok")
}

func getUser(w http.ResponseWriter, r *http.Request){
  content, err := ioutil.ReadAll(r.Body)
  if err != nil {
    WriteStatus(w,500,"api user get error")
    return
  }
  userGetReq:=UserGetReq{}
  err = json.Unmarshal(content, &userGetReq) 
  if err != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api user get data error")
    return
  }


  log.Printf("userGet{%s}",userGetReq.Email)

  _,res:=GetKV([]byte(userGetReq.Email),[]byte("users"))
  if res != nil {

    user := &User{}
    err = user.GobDecode(res)
    if err!=nil{
      log.Printf("err: %s",err)
      WriteStatus(w,500,"api user get decode error")
      return
    }
/*
    userGetResp := &UserGetResp{
      Id:user.Id,
      Created:user.Created,
      Email:user.Email,
    }
*/
    w.WriteHeader(200)
    err = WriteJSON(w,user)
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
  var err error
  db, err = bolt.Open("data.db", 0600, nil)
  if err != nil {
    log.Fatal(err)
  }
  err = db.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucketIfNotExists([]byte("users"))
    if err != nil {
        return fmt.Errorf("create bucket error: %s", err)
    }
    return nil
  })  
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()
  log.Println("http.ListenAndServe:8080");
  http.ListenAndServe(":8080", &Faas{})
}
