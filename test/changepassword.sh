curl -v -XPOST -H "Content-Type: application/json" -d '{"email":"kosmodb@gmail.com","oldpassword":"123456789","newpassword":"abcdefg"}' http://localhost:8000/api/v1/users/password/change
