package faas

import (
  "log"
  "net/http"
//  "bytes"
//  "encoding/gob"
//  "time"  
  "io/ioutil"
//  "encoding/json"
)


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
func PostTestKV(w http.ResponseWriter, r *http.Request, bucket string, key string){
  /*
  var bucket,key string
  w.Header().Set("Content-Type", "application/json")
  if bucket = r.URL.Query().Get("bucket"); bucket == "" {
    log.Printf("postTestKV /api/v1/set bucket error")
    http.Error(w,"{\"status\" : \"bucket error\"}",http.StatusBadRequest)
    return
  }
  if key = r.URL.Query().Get("key"); key == "" {
    log.Printf("postTestKV /api/v1/set key error")
    http.Error(w,"{\"status\" : \"key error\"}",http.StatusBadRequest)
    return
  }
  */
  err := UpdBucket([]byte(bucket))
  if err != nil {
    log.Printf("postTestKV create bucket %s error %v",bucket,err)
    http.Error(w,"{\"status\" : \"create bucket error\"}",http.StatusBadRequest)
    return
  }
  var value []byte  
  if value, err = ioutil.ReadAll(r.Body); err != nil {
    log.Printf("postTestKV read error %v",err)
    http.Error(w,"{\"status\" : \"body error\"}",http.StatusBadRequest)
    return
  }  
  if err=UpdKV([]byte(key),[]byte(value),[]byte(bucket)); err!=nil{
    log.Printf("postTestKV update error %v",err)
    http.Error(w,"{\"status\" : \"update error\"}",http.StatusBadRequest)
    return
  }
  w.WriteHeader(http.StatusOK);w.Write([]byte("{\"status\" : \"ok\"}"))
}


// curl -v -XGET 'http://localhost:8080/api/v1/get?bucket=mydata&key=mykey'
func GetTestKV(w http.ResponseWriter, r *http.Request, bucket string, key string){
  /*
  var bucket,key string
  if bucket = r.URL.Query().Get("bucket"); bucket == "" {
    log.Printf("getTestKV /api/v1/get bucket error")
    http.Error(w,"{\"status\" : \"bucket error\"}",http.StatusBadRequest)
    return
  }
  if key = r.URL.Query().Get("key"); key == "" {
    log.Printf("getTestKV /api/v1/get key error")
    http.Error(w, "{\"status\" : \"key error\"}", http.StatusBadRequest)
    return
  }
  */
  var err error
  var value []byte  
  err,value=GetKV([]byte(key),[]byte(bucket)); 
  if err != nil {
    log.Printf("getTestKV /api/v1/get value error %v",err)
    http.Error(w,"{\"status\" : \"error\"}",http.StatusBadRequest)
    return
  }
  if value == nil {
    log.Printf("getTestKV /api/v1/get value nil")
    http.Error(w,"{\"status\" : \"not found\"}",http.StatusNotFound)
    return
  }
    w.WriteHeader(http.StatusOK)
    w.Write(value) 
}

