package faas

import (
	"log"
	"net/http"
	//  "bytes"
	//  "encoding/gob"
	//  "time"
	"io/ioutil"
	//  "encoding/json"
	//  "crypto/rand"
	"fmt"
)

// https://en.wikipedia.org/wiki/List_of_HTTP_status_codes

// curl -v -XPOST -d '123456789abcdef' http://localhost:8080/api/v1/obj/mydata/mykey
func PostObjKV(w http.ResponseWriter, r *http.Request, bucket string, key string) {
	w.Header().Set("Content-Type", "application/json")
	err := UpdBucket([]byte(bucket))
	if err != nil {
		log.Printf("PostObjKV create bucket %s error %v", bucket, err)
		http.Error(w, "{\"status\" : \"create bucket error\"}", http.StatusBadRequest)
		return
	}
	var value []byte
	if value, err = ioutil.ReadAll(r.Body); err != nil {
		log.Printf("PostObjKV read error %v", err)
		http.Error(w, "{\"status\" : \"body error\"}", http.StatusBadRequest)
		return
	}

	if key == "" {
		key = NewUUID()
	} else {
		e, v := GetKV([]byte(key), []byte(bucket))
		if e == nil && v != nil {
			log.Printf("PostObjKV api/v1/obj/%s already exists", key)
			http.Error(w, "{\"status\" : \"already exists\"}", http.StatusBadRequest)
			return
		}
	}

	if err = UpdKV([]byte(key), []byte(value), []byte(bucket)); err != nil {
		log.Printf("PostObjKV update error %v", err)
		http.Error(w, "{\"status\" : \"update error\"}", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\n\"id\":\"%s\",\n\"status\":\"ok\"\n}", key)))
}

// curl -v -XPUT -d '123456789abcdef' http://localhost:8080/api/v1/obj/mydata/mykey
func PutObjKV(w http.ResponseWriter, r *http.Request, bucket string, key string) {
	w.Header().Set("Content-Type", "application/json")
	if len(bucket) == 0 || len(key) == 0 {
		log.Printf("PutObjKV params error")
		http.Error(w, "{\"status\" : \"error\"}", http.StatusBadRequest)
		return
	}

	err := UpdBucket([]byte(bucket))
	if err != nil {
		log.Printf("PutObjKV UpdBucket %s error %v", bucket, err)
		http.Error(w, "{\"status\" : \"bucket error\"}", http.StatusBadRequest)
		return
	}
	var value []byte
	if value, err = ioutil.ReadAll(r.Body); err != nil {
		log.Printf("PutObjKV body error %v", err)
		http.Error(w, "{\"status\" : \"body error\"}", http.StatusBadRequest)
		return
	}

	e, v := GetKV([]byte(key), []byte(bucket))
	if e != nil || v == nil {
		log.Printf("PutObjKV api/v1/obj/%s don't exists", key)
		http.Error(w, "{\"status\" : \"don't exists\"}", http.StatusBadRequest)
		return
	}

	if err = UpdKV([]byte(key), []byte(value), []byte(bucket)); err != nil {
		log.Printf("PutObjKV update error %v", err)
		http.Error(w, "{\"status\" : \"update error\"}", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\n\"id\":\"%s\",\n\"status\":\"ok\"\n}", key)))
}

// curl -v -XGET 'http://localhost:8080/api/v1/obj/mydata/mykey'
func GetObjKV(w http.ResponseWriter, r *http.Request, bucket string, key string) {
	var err error
	var value []byte
	err, value = GetKV([]byte(key), []byte(bucket))
	if err != nil {
		log.Printf("GetObjKV /api/v1/get value error %v", err)
		http.Error(w, "{\"status\" : \"error\"}", http.StatusBadRequest)
		return
	}
	if value == nil {
		log.Printf("getTestKV /api/v1/get value nil")
		http.Error(w, "{\"status\" : \"not found\"}", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(value)
}

// curl -v -XDELETE 'http://localhost:8080/api/v1/obj/mydata/mykey'
func DelObjKV(w http.ResponseWriter, r *http.Request, bucket string, key string) {
	if err := DelKV([]byte(key), []byte(bucket)); err != nil {
		log.Printf("DelObjKV /api/v1/get value error %v", err)
		http.Error(w, "{\"status\" : \"error\"}", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n\"status\":\"ok\"\n}"))
}
