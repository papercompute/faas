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

    // objects

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


    // users

    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users/register
    router.POST("/api/v1/users/register",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      RegisterNewUser(w,r)
    })
    // curl -v -XGET  http://localhost:8080/api/v1/users/confirm/email/3bea3a7ba0814591852016fdc8c3ecce
    router.GET("/api/v1/users/confirm/email/:userid",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      ConfirmUserEmail(w,r,ps.ByName("userid"))
    })
    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com"}' http://localhost:8080/api/v1/users/resend/confirm/email
    router.POST("/api/v1/users/resend/confirm/email",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      ResendUserConfirmationEmail(w,r)
    })
    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789"}' http://localhost:8080/api/v1/users/login
    router.POST("/api/v1/users/login",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      LoginUser(w,r)
    })
    // curl -v -XGET -H "X-Auth-Token: 213ba332342234234920349282882822" http://localhost:8080/api/v1/users/refresh
    router.GET("/api/v1/users/refresh",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      RefreshUserToken(w,r)
    })
    // curl -v -XGET -H "Content-Type: application/json" -H "X-Auth-Token: 213ba332342234234920349282882822" http://localhost:8080/api/v1/users/info
    router.GET("/api/v1/users/info",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      GetUserInfo(w,r)
    })
    // curl -v -XGET -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com"}' http://localhost:8080/api/v1/users/password/reset/token
    router.GET("/api/v1/users/password/reset/token",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      SendPasswordResetTokenToUserEmail(w,r)
    })
    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","password":"123456789","token":"3bea3a7ba0814591852016fdc8c3ecce"}' 
    // http://localhost:8080/api/v1/users/password/reset
    router.POST("/api/v1/users/password/reset",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      ResetUserPasswordWithNewOneByToken(w,r)
    })
    // curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"sobaka@drug.com","oldpassword":"123456789","newpassword":"123456789"}' 
    // http://localhost:8080/api/v1/users/password/change
    router.POST("/api/v1/users/password/change",  func (w http.ResponseWriter, r *http.Request, ps httprouter.Params){
      ChangeUserPassword(w,r)
    })


    if public_dir!=""{
    	router.NotFound = http.FileServer(http.Dir(public_dir))
	  }

    return http.ListenAndServe(host, router)

}
