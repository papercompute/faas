package faas

import (
//    "log"
    "net/smtp"
)


func SendMail(toEmail, subj, body string) error {

 auth := smtp.PlainAuth("", CFG.Email, CFG.EmailPassword, "smtp.gmail.com")

  // Connect to the server, authenticate, set the sender and recipient,
  // and send the email all in one step.
  to := []string{toEmail}
  msg := []byte("To: "+toEmail+"\r\n" +"Subject: "+subj+"\r\n\r\n"+body+"\r\n")
  return smtp.SendMail("smtp.gmail.com:587", auth, CFG.Email, to, msg)
}
