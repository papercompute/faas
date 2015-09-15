package faas

import (
//  "log"
  "net/http"
//  "bytes"
//  "encoding/gob"
//  "time"  
//  "io/ioutil"
//  "encoding/json"
//  "crypto/rand"
//  "fmt"
  "github.com/julienschmidt/httprouter"
)

func ListenAndServe(host string, public_dir string) error {
    router := httprouter.New()

    // curl -v -XPOST -d '123456789abcdef!' 'http://localhost:8080/api/v1/obj/mydata'
    router.POST("/api/v1/obj/:bucket", func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      PostObjKV(w,r,ps.ByName("bucket"),"")
    })  
    // curl -v -XPOST -d '123456789abcdef!' 'http://localhost:8080/api/v1/obj/mydata/mykey'
    router.POST("/api/v1/obj/:bucket/:key", func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      PostObjKV(w,r,ps.ByName("bucket"),ps.ByName("key"))
    })  
    // curl -v -XPUT -d '123456789abcdef!' 'http://localhost:8080/api/v1/obj/mydata/mykey'
    router.PUT("/api/v1/obj/:bucket/:key", func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      PutObjKV(w,r,ps.ByName("bucket"),ps.ByName("key"))
    })  
    // curl -v -XGET 'http://localhost:8080/api/v1/obj/mydata/mykey'
    router.GET("/api/v1/obj/:bucket/:key",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      GetObjKV(w,r,ps.ByName("bucket"),ps.ByName("key"))
    })
    // curl -v -XDELETE 'http://localhost:8080/api/v1/obj/mydata/mykey'
    router.DELETE("/api/v1/obj/:bucket/:key",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      DelObjKV(w,r,ps.ByName("bucket"),ps.ByName("key"))
    })

    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users/login
    router.POST("/api/v1/users/login",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      LoginUser(w,r)
    })
    // curl -v -XPOST -H "Content-Type: application/json" -d '{"token":"4550aa3ef1584a649a6dcc6a151762cb"}' http://localhost:8080/api/v1/users/logout
    router.POST("/api/v1/users/logout",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      LogoutUser(w,r)
    })

    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users
    router.POST("/api/v1/users",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      PostNewUser(w,r)
    })
    // curl -v -XGET -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com"}' http://localhost:8080/api/v1/users'
    router.GET("/api/v1/users",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      GetUser(w,r)
    })

    if public_dir!=""{
    	router.NotFound = http.FileServer(http.Dir(public_dir)).ServeHTTP
	  }

    return http.ListenAndServe(host, router)

}
