package faas

import (
  "log"
  "net/http"
  "encoding/json"
  "crypto/md5"
  "encoding/hex"  
  "io/ioutil"  
)



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
