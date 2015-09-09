package faas

import (
  "log"
//  "strings"
  "github.com/boltdb/bolt"

)



var db *bolt.DB

func UpdBucket(bucket []byte) error {
  err := db.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucketIfNotExists(bucket)
    return err
  })
  return err
}

func UpdKV(key []byte, value []byte, bucket []byte) error {
  return db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket(bucket)
    return b.Put(key, value)
  })
}
func GetKV(key []byte, bucket []byte) (error, []byte) {
  var result []byte = nil
  err := db.View(func(tx *bolt.Tx) error {
    if b := tx.Bucket(bucket); b != nil {
      result = b.Get(key)
    }  
    return nil
  })
  return err, result
}

func InitDB2(dbname *string, buckets []string){
  var err error 
  db, err = bolt.Open(*dbname, 0600, nil)
  if err != nil {
    log.Fatal(err)
  }
  err = db.Update(func(tx *bolt.Tx) error {
  	for _,bucket:=range buckets{
      _, err := tx.CreateBucketIfNotExists([]byte(bucket))
      if err != nil {
        log.Fatal(err)
      }
    }
    return nil
  })  
  if err != nil {
    log.Fatal(err)
  }
}

func FinitDB2(){
  db.Close()
}