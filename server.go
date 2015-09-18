package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
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
	"./lib"
	"strings"
  //"golang.org/x/crypto/bcrypt"
)

/*
example data config
{

"email" : "kat@kats.com"
"email_password" : "katsecret"

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



type Faas struct{}
type Storage struct{}


func (e *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
	http.NotFound(w, r)
}


func main() {

  var err error


	var cfg_file_name *string = flag.String("cfg", "config.json", "config file")
	var db_file_name *string = flag.String("db", "data.db", "db file")
	var cfg_id *int = flag.Int("id", 0, "id")

	
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

	err = json.Unmarshal(cfg_file, &faas.CFG)
	if err != nil {
		fmt.Printf("Cfg file json error: %v\n", err)
		os.Exit(1)
	}
	faas.CFG.Id = uint(*cfg_id)
	faas.CFG.Stora_host = faas.CFG.Hosts[faas.CFG.Id*2]
	faas.CFG.Web_host = faas.CFG.Hosts[faas.CFG.Id*2+1]
  faas.CFG.Url = faas.CFG.Web_host

	fmt.Printf("id: %d\n", *cfg_id)
	fmt.Printf("replica: %d\n", faas.CFG.Replica)
	fmt.Printf("cfg: %s,%s\n", faas.CFG.Stora_host, faas.CFG.Web_host)

	faas.InitDB2(db_file_name, 
    []string{
      faas.BucketTests, 
      faas.BucketUsers, 
      faas.BucketAwaitEmailConfirmationIds,
      faas.BucketPasswordResetIds,
      })
	defer faas.FinitDB2()


    
    err=faas.TestIDB()

	log.Fatal("Hipster! %v",err)


	finish := make(chan bool)

  go func() {
    log.Println("http.ListenAndServe Stora_host "+faas.CFG.Stora_host);
    h:=strings.Split(faas.CFG.Stora_host, ":")  
    err:=http.ListenAndServe(":"+h[1], &Storage{})
    if err != nil {
      log.Fatal("ListenAndServe: Stora_host ", err)
    }
  }()

  go func() {    
    log.Println("http.ListenAndServe Web_host "+faas.CFG.Web_host);    
    h:=strings.Split(faas.CFG.Web_host, ":")
    err:=faas.ListenAndServe(":"+h[1],"public")
    if err != nil {
      log.Fatal("ListenAndServe: Web_host ", err)
    }
  }()

	<-finish
}
