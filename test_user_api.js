var http = require('http');


function tryToRegisterNewUser(email,password){

var postData = JSON.stringify({
  "email":email,
  "password":password
});

var options = {
  hostname: 'localhost',
  port: 8000,
  path: '/api/v1/users/register',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Content-Length': postData.length
  }
};

var req = http.request(options, function(res) {
  console.log('STATUS: ' + res.statusCode);
  console.log('HEADERS: ' + JSON.stringify(res.headers));
  res.setEncoding('utf8');
  body="";
  res.on('data', function (chunk) {
  	body+=chunk;
  });
  res.on('end', function() {
    console.log('BODY: ' + body);
    if(res.statusCode == 201){
    	js=JSON.parse(body);
    	//getUserInfo(js.id);
      console.log(js)
    }
  })
});

req.on('error', function(e) {
  console.log('problem with request: ' + e.message);
});

req.write(postData);
req.end();

}


function getUserInfo(id,token){

var getData = JSON.stringify({
  "id":id
});

var options = {
  hostname: 'localhost',
  port: 8000,
  path: '/api/v1/users',
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
    'X-Auth-Token': token,    
    'Content-Length': getData.length
  }
};

var req = http.request(options, function(res) {
  console.log('STATUS: ' + res.statusCode);
  console.log('HEADERS: ' + JSON.stringify(res.headers));
  res.setEncoding('utf8');
  body="";
  res.on('data', function (chunk) {
  	body+=chunk;
  });
  res.on('end', function() {
    console.log('BODY: ' + body);
    if(res.statusCode == 200){
    	js=JSON.parse(body);
    	console.log('TEST OK');    	
    }
  })
});

req.on('error', function(e) {
  console.log('problem with request: ' + e.message);
});

req.write(getData);
req.end();
}



function loginUser(email,password){

var loginData = JSON.stringify({
  "email":email,
  "password":password
});

var options = {
  hostname: 'localhost',
  port: 8000,
  path: '/api/v1/users/login',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Content-Length': loginData.length
  }
};

var req = http.request(options, function(res) {
  console.log('STATUS: ' + res.statusCode);
  console.log('HEADERS: ' + JSON.stringify(res.headers));
  res.setEncoding('utf8');
  body="";
  res.on('data', function (chunk) {
  	body+=chunk;
  });
  res.on('end', function() {
    console.log('BODY: ' + body);
    if(res.statusCode == 200){
    	js=JSON.parse(body);
    	//getUserInfo(js.id);
    	console.log("token:",js.token);

    	//getUserInfo(js.id,js.token);
    }
  })
});

req.on('error', function(e) {
  console.log('problem with request: ' + e.message);
});

req.write(loginData);
req.end();

}





// start
//tryToRegisterNewUser("kosmodb@gmail.com","123456789");
loginUser("kosmodb@gmail.com","123456789");
//postNewUser("herome@qwer.com","12312312312312");
//loginUser("herome@qwer.com","12312312312312");
//loginUser("asda@asdasd.com","adj2io3e23b23ek2b3ekj2b3ek2b3ekj23");

