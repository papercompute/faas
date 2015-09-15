package main

import (
  "flag"
  "os"
	"net/http"
	"log"
	"runtime"
  "fmt"
//"crypto/md5"
//    "encoding/hex"
//	"io"
  "io/ioutil"
//  "net/url"
  "encoding/json"
//  "encoding/gob"
//  "time"
//	"regexp"
//  "bytes"
//  "runtime"
//  "github.com/julienschmidt/httprouter"
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

type Faas struct {}
type Storage struct {}




func (e *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
  http.NotFound(w, r)
}


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

  faas.InitDB2(db_file_name,[]string{"users","tests","usersids"})
  defer faas.FinitDB2()

// server setup
  finish := make(chan bool)


  go func() {
    log.Println("http.ListenAndServe Stora_host "+cfg.Stora_host);
    h:=strings.Split(cfg.Stora_host, ":")  
    err:=http.ListenAndServe(":"+h[1], &Storage{})
    if err != nil {
      log.Fatal("ListenAndServe: Stora_host ", err)
    }
  }()

  go func() {    
    log.Println("http.ListenAndServe Web_host "+cfg.Web_host);    
    h:=strings.Split(cfg.Web_host, ":")
    err:=faas.ListenAndServe(":"+h[1],"public")
    if err != nil {
      log.Fatal("ListenAndServe: Web_host ", err)
    }
  }()

  <-finish
}
