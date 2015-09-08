package main

import (
  "flag"
  "os"
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
  "bytes"
//  "runtime"
  "strings"
  "./lib"
)


/*
example data config
{  
"hosts" : [
  "127.0.0.1:9000",  "127.0.0.1:8000",
  "127.0.0.1:9001",  "127.0.0.1:8001",
  "127.0.0.1:9002",  "127.0.0.1:8002",

  "127.0.0.1:9003",  "127.0.0.1:8003",
  "127.0.0.1:9004",  "127.0.0.1:8004",
  "127.0.0.1:9005",  "127.0.0.1:8005",

  "127.0.0.1:9006",  "127.0.0.1:8006",
  "127.0.0.1:9007",  "127.0.0.1:8007",
  "127.0.0.1:9008",  "127.0.0.1:8008",

  "127.0.0.1:9009",  "127.0.0.1:8009",
  "127.0.0.1:9010",  "127.0.0.1:8010",
  "127.0.0.1:9011",  "127.0.0.1:8011"
],
"replica" : "3"
}
*/

type Config struct {
  // key space (0..2^64)
  Hosts []string  `json:"hosts"`
  Replica uint  `json:"replica"`
  Id uint  `json:"-"`  
  Stora_host string  `json:"-"`  
  Web_host string  `json:"-"`  
  // hosts[(id % len(hosts))*replica]
}

var cfg Config
/*
func (c *Config) getUrlfromID(uint64 id)[]string {
  i:=id%N
}
*/




func GetMD5Hash(text string) string {
    hasher := md5.New()
    hasher.Write([]byte(text))
    return hex.EncodeToString(hasher.Sum(nil))
}

// http.Error(rw, err.Error(), 500)

func WriteStatus(w http.ResponseWriter, code int, status string) {
    w.WriteHeader(code)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte("{\"status\" : \""+status+"\"}"))
}

func WriteJSON(w http.ResponseWriter, v interface{}) error {
  b, err := json.MarshalIndent(v, "", "  ")
  if err != nil {
    log.Printf("WriteJSON error %v",err)
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
type Storage struct {}

type User struct {
  Id uint64 `json:"id"`
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




type NewUserPostReq struct {
  Email string `json:"email"`
  Password string `json:"password"`
}
type UserGetReq struct {
  Email string `json:"email"`
}
type UserGetResp struct {
  Id uint64 `json:"id"`
  Created time.Time `json:"created"`
  Email string `json:"email"`
}

var userIdCounter uint64 = 127 

func GetUserId() uint64{
  uid := userIdCounter
  userIdCounter = userIdCounter + 1
  return uid  
}


// curl -v -XPOST -H "Content-Type: application/json" \
// -d '{"email":"sobaka@drug.com","password":"123456789"}' \
//'http://localhost:8080/api/v1/newuser'
func postNewUser(w http.ResponseWriter, r *http.Request){
  userPost:=NewUserPostReq{}
  var err error
  if err = ReadJSON(r, &userPost); err != nil {
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

  var res []byte
  err,res=db2.GetKV([]byte(userPost.Email),[]byte("users"))
  if res != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api newuser user already exists")
    return
  }

  user := &User{
    Id:GetUserId(),
    Created: time.Now(), //int32(time.Now().Unix()),
    Email:userPost.Email,
    PasswordHash:GetMD5Hash(userPost.Password),
  }

  var encoded []byte
  encoded, err = user.GobEncode()
  if err != nil {
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api newuser user error")
    return
  }


  err=db2.UpdKV([]byte(userPost.Email),encoded,[]byte("users"))
  if err==nil{
    log.Printf("err: %s",err)
    WriteStatus(w,500,"api newuser update error")
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

  _,res:=db2.GetKV([]byte(userGetReq.Email),[]byte("users"))
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


/*
func DecodeJson(r *http.Request, v interface{}) error {
  content, err := ioutil.ReadAll(r.Body)
  if err != nil {
    return err
  }
  err = json.Unmarshal(content, v)
  if err != nil {
    return err
  }
  return nil
}*/

// https://en.wikipedia.org/wiki/List_of_HTTP_status_codes


// curl -v -XPOST -d '123456789abcdef' 'http://localhost:8080/api/v1/set?bucket=mydata&key=mykey'
func postTestKV(w http.ResponseWriter, r *http.Request){
  var bucket,key string
  w.Header().Set("Content-Type", "application/json")
  if bucket = r.URL.Query().Get("bucket"); bucket == "" {
    log.Printf("postTestKV /api/v1/set bucket error")
    http.Error(w,"{\"status\" : \"bucket error\"}",400)
    return
  }
  if key = r.URL.Query().Get("key"); key == "" {
    log.Printf("postTestKV /api/v1/set key error")
    http.Error(w,"{\"status\" : \"key error\"}",400)
    return
  }
  err := db2.UpdBucket([]byte(bucket))
  if err != nil {
    log.Printf("postTestKV create bucket %s error %v",bucket,err)
    http.Error(w,"{\"status\" : \"create bucket error\"}",400)
    return
  }
  var value []byte  
  if value, err = ioutil.ReadAll(r.Body); err != nil {
    log.Printf("postTestKV read error %v",err)
    http.Error(w,"{\"status\" : \"body error\"}",400)
    return
  }  
  if err=db2.UpdKV([]byte(key),[]byte(value),[]byte(bucket)); err!=nil{
    log.Printf("postTestKV update error %v",err)
    http.Error(w,"{\"status\" : \"update error\"}",400)
    return
  }
  w.WriteHeader(200);w.Write([]byte("{\"status\" : \"ok\"}"))
}


// curl -v -XGET 'http://localhost:8080/api/v1/get?bucket=mydata&key=mykey'
func getTestKV(w http.ResponseWriter, r *http.Request){
  var bucket,key string
  if bucket = r.URL.Query().Get("bucket"); bucket == "" {
    log.Printf("getTestKV /api/v1/get bucket error")
    http.Error(w,"{\"status\" : \"bucket error\"}",400)
    return
  }
  if key = r.URL.Query().Get("key"); key == "" {
    log.Printf("getTestKV /api/v1/get key error")
    http.Error(w, "{\"status\" : \"key error\"}", 400)
    return
  }
  var err error
  var value []byte  
  if err,value=db2.GetKV([]byte(key),[]byte(bucket)); err != nil || value == nil {
    log.Printf("getTestKV /api/v1/get key error")
    http.Error(w,"{\"status\" : \"value error\"}",400)
    return
  }
    w.WriteHeader(200)
    w.Write(value) 
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
    case "/api/v1/set":
     if r.Method == "POST" {
       postTestKV(w,r)
       return
     }
    case "/api/v1/get":
     if r.Method == "GET" {
       getTestKV(w,r)
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



func (e *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
  http.NotFound(w, r)
}


/*

  router := httprouter.New()
  router.NotFound = http.FileServer(http.Dir("public")).ServeHTTP
  log.Fatal(http.ListenAndServe(":8000", router))
*/

func main() {


  var cfg_file_name* string = flag.String("cfg", "config.json", "config file")  
  var db_file_name* string = flag.String("db", "data.db", "db file")  
  var cfg_id* int = flag.Int("id", 0, "id")  

  var err error
  runtime.GOMAXPROCS(runtime.NumCPU() - 1)
//  runtime.GOMAXPROCS(1)


// read config
  cfg_file, e := ioutil.ReadFile(*cfg_file_name)
  if e != nil {
    fmt.Printf("File error: %v\n", e)
    os.Exit(1)
  }
  fmt.Printf("%s\n", string(cfg_file))

  flag.Parse()

  err = json.Unmarshal(cfg_file, &cfg)
  if err != nil {
    fmt.Printf("Cfg file json error: %v\n", err)
    os.Exit(1)
  }
  cfg.Id = uint(*cfg_id)
  cfg.Stora_host = cfg.Hosts[cfg.Id*2];
  cfg.Web_host = cfg.Hosts[cfg.Id*2+1];

  fmt.Printf("id: %d\n", *cfg_id)
  fmt.Printf("replica: %d\n", cfg.Replica)
  fmt.Printf("cfg: %s,%s\n", cfg.Stora_host,cfg.Web_host)

  db2.InitDB2(db_file_name,[]string{"users","tests"})
  defer db2.FinitDB2()

// server setup
  finish := make(chan bool)

  go func() {
    log.Println("http.ListenAndServe Stora_host "+cfg.Stora_host);
    h:=strings.Split(cfg.Stora_host, ":")
    http.ListenAndServe(":"+h[1], &Storage{})
  }()

  go func() {
    log.Println("http.ListenAndServe Web_host "+cfg.Web_host);
    h:=strings.Split(cfg.Web_host, ":")
    http.ListenAndServe(":"+h[1], &Faas{})
  }()

  <-finish
}
