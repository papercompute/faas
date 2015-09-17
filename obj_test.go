package faas

import (
//	"./lib"
//	"time"
	"net/http"
	"strings"
//	"fmt"
	"io/ioutil"
	"encoding/json"
	"testing"
)

func TestObjects(t *testing.T) {
/*
   webHost:="127.0.0.1:8000"

  go func() {    
    log.Println("http.ListenAndServe Web_host "+webHost);    
    h:=strings.Split(webHost, ":")
    err:=faas.ListenAndServe(":"+h[1],"public")
    if err != nil {
      log.Fatal("ListenAndServe: "+ webHost + " "+err)
    }
  }()

  time.Sleep(100 * time.Millisecond)
*/
  var url = "http://127.0.0.1:8000"

  var data string = `1234567890abcdef!@#$%^&*()_+`

  // create object

  r,err:=http.Post(url+"/api/v1/obj/testbucket", "application/json", 
  	strings.NewReader(data))

  if err!=nil {
  	t.Fatal(err)
  }

  if r.StatusCode!=200{
  	t.Fatal(err)
  }

  var content []byte  
  if content, err = ioutil.ReadAll(r.Body); err != nil {
  	t.Fatal(err)
  }

  jsmap:=map[string]string{}

  if err = json.Unmarshal(content, &jsmap); err != nil {
  	t.Fatal(err)
  }

  if jsmap["status"]!="ok"{
  	t.Fatal(err)
  }
  if len(jsmap["id"])!=32{
  	t.Fatal(err)
  }

  t.Log(jsmap["id"],jsmap["status"])

  // get object


  r,err=http.Get(url+"/api/v1/obj/testbucket/"+jsmap["id"])

  if err!=nil {
  	t.Fatal(err)
  }

  if r.StatusCode!=200{
  	t.Fatal(err)
  }

  //var content []byte  
  if content, err = ioutil.ReadAll(r.Body); err != nil {
  	t.Fatal(err)
  }

  if string(content)!=data{
  	t.Fatal(err)
  }

  t.Log(string(content))


  // delete object

  req,err1:=http.NewRequest("DELETE", url+"/api/v1/obj/testbucket/"+jsmap["id"], nil)

  if err1!=nil {
  	t.Fatal(err1)
  }

  r,err= http.DefaultClient.Do(req)

  if err!=nil {
  	t.Fatal(err)
  }

  if r.StatusCode!=200{
  	t.Fatal(err)
  }

  if content, err = ioutil.ReadAll(r.Body); err != nil {
  	t.Fatal(err)
  }


  jsmap:=map[string]string{}

  if err = json.Unmarshal(content, &jsmap); err != nil {
  	t.Fatal(err)
  }

  if jsmap["status"]!="ok"{
  	t.Fatal(err)
  }
  
  t.Log(jsmap["status"])


}
