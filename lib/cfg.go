package faas

import (
  //  "log"
  // "net/http"
  //  "bytes"
  //  "encoding/gob"
  //"time"
  //  "io/ioutil"
  //  "encoding/json"
  //"fmt"

)



type Config struct {
  Hosts      []string `json:"hosts"`
  Replica    uint     `json:"replica"`
  Id         uint     `json:"-"`
  Stora_host string   `json:"-"`
  Web_host   string   `json:"-"`
  // hosts[(id % len(hosts))*replica]
  Url        string   `json:"-"`
  Email      string   `json:"email"`
  EmailPassword   string   `json:"email_password"`
}

var CFG Config
