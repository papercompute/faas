package faas

import (
  "log"
  "net/http"
  "encoding/json"
  "crypto/md5"
  "encoding/hex"  
  "io/ioutil"  
  "crypto/rand"
  "fmt"

)


type UUID [16]byte

func NewUUID() *UUID {
  u := &UUID{}
  _, err := rand.Read(u[:16])
  if err != nil {
    panic(err)
  }

  u[8] = (u[8] | 0x80) & 0xBf
  u[6] = (u[6] | 0x40) & 0x4f
  return u
}

func (u *UUID) String() string {
//  return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
    return fmt.Sprintf("%x%x%x%x", u[:4],u[4:8],u[8:12],u[12:])
}

/*
type UUID8 [8]byte

func NewUUID8() *UUID8 {
  u := &UUID8{}
  _, err := rand.Read(u[:8])
  if err != nil {
    panic(err)
  }

//  u[8] = (u[8] | 0x80) & 0xBf
  u[6] = (u[6] | 0x40) & 0x4f
  return u
}

func (u *UUID8) String() string {
//  return fmt.Sprintf("%x-%x-%x-%x-%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
    return fmt.Sprintf("%x%x", u[:4],u[4:8])
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
    log.Printf("WriteJSON json.MarshalIndent error %v",err)
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
