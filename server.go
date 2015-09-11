package main

import (
  "flag"
  "os"
	"net/http"
	"log"
	"runtime"
  "fmt"
//"crypto/md5"
//    "encoding/hex"
//	"io"
  "io/ioutil"
//  "net/url"
  "encoding/json"
//  "encoding/gob"
//  "time"
//	"regexp"
//  "bytes"
//  "runtime"
//  "github.com/julienschmidt/httprouter"
  "strings"
  "./lib"
)


/*
example data config
{  
"hosts" : [
  "127.0.0.1:9000",  "127.0.0.1:8000",
  "127.0.0.1:9001",  "127.0.0.1:8001",
  "127.0.0.1:9002",  "127.0.0.1:8002",

  "127.0.0.1:9003",  "127.0.0.1:8003",
  "127.0.0.1:9004",  "127.0.0.1:8004",
  "127.0.0.1:9005",  "127.0.0.1:8005",

  "127.0.0.1:9006",  "127.0.0.1:8006",
  "127.0.0.1:9007",  "127.0.0.1:8007",
  "127.0.0.1:9008",  "127.0.0.1:8008",

  "127.0.0.1:9009",  "127.0.0.1:8009",
  "127.0.0.1:9010",  "127.0.0.1:8010",
  "127.0.0.1:9011",  "127.0.0.1:8011"
],
"replica" : "3"
}
*/

type Config struct {
  // key space (0..2^64)
  Hosts []string  `json:"hosts"`
  Replica uint  `json:"replica"`
  Id uint  `json:"-"`  
  Stora_host string  `json:"-"`  
  Web_host string  `json:"-"`  
  // hosts[(id % len(hosts))*replica]
}

var cfg Config
/*
func (c *Config) getUrlfromID(uint64 id)[]string {
  i:=id%N
}
*/





type Faas struct {}
type Storage struct {}


/*
routes := map[string]Handle{
  "/api/v1/newuser":faas.PostNewUser,
  "/api/v1/user":faas.PostNewUser,
  "/api/v1/set":faas.PostTestKV,
  "/api/v1/get":faas.GetTestKV,
}
*/

// royhttp://vluxe.io/golang-router.html

/*
/// Handle is just like "net/http" Handlers, only takes params.
type Handle func(http.ResponseWriter, *http.Request, url.Values)

// Router name says it all.
type Router struct {
  tree        *node
  rootHandler Handle
}

// New creates a new router. It takes the root (fall through) route
// like how the default mux works. The only difference, you get to specify one.
func New(rootHandler Handle) *Router {
  node := node{component: "/", isNamedParam: false, methods: make(map[string]Handle)}
  return &Router{tree: &node, rootHandler: rootHandler}
}

// Handle takes an http handler, method, and pattern for a route.
func (r *Router) Handle(method, path string, handler Handle) {
  if path[0] != '/' {
    panic("Path has to start with a /.")
  }
  r.tree.addNode(method, path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  req.ParseForm()
  params := req.Form
  node, _ := r.tree.traverse(strings.Split(req.URL.Path, "/")[1:], params)
  if handler := node.methods[req.Method]; handler != nil {
    handler(w, req, params)
  } else {
    r.rootHandler(w, req, params)
  }
}

// node represents a struct of each node in the tree.
type node struct {
  children     []*node
  component    string
  isNamedParam bool
  methods      map[string]Handle
}

// addNode - adds a node to our tree. Will add multiple nodes if path
// can be broken up into multiple components. Those nodes will have no
// handler implemented and will fall through to the default handler.
func (n *node) addNode(method, path string, handler Handle) {
  components := strings.Split(path, "/")[1:]
  count := len(components)

  for {
    aNode, component := n.traverse(components, nil)
    if aNode.component == component && count == 1 { // update an existing node.
      aNode.methods[method] = handler
      return
    }
    newNode := node{component: component, isNamedParam: false, methods: make(map[string]Handle)}

    if len(component) > 0 && component[0] == ':' { // check if it is a named param.
      newNode.isNamedParam = true
    }
    if count == 1 { // this is the last component of the url resource, so it gets the handler.
      newNode.methods[method] = handler
    }
    aNode.children = append(aNode.children, &newNode)
    count--
    if count == 0 {
      break
    }
  }
}

// traverse moves along the tree adding named params as it comes and across them.
// Returns the node and component found.
func (n *node) traverse(components []string, params url.Values) (*node, string) {
  component := components[0]
  if len(n.children) > 0 { // no children, then bail out.
    for _, child := range n.children {
      if component == child.component || child.isNamedParam {
        if child.isNamedParam && params != nil {
          params.Add(child.component[1:], component)
        }
        next := components[1:]
        if len(next) > 0 { // http://xkcd.com/1270/
          return child.traverse(next, params) // tail recursion is it's own reward.
        } else {
          return child, component
        }
      }
    }
  }
  return n, component
}

*/

/*

var fileServer = http.FileServer(http.Dir("public"))

func (e *Faas) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
  switch r.URL.Path {

    case "/api/v1/newuser":
     if r.Method == "POST" {
       faas.PostNewUser(w,r)
       return
     }
    case "/api/v1/user":
     if r.Method == "GET" {
       faas.GetUser(w,r)
       return
     }
    case "/api/v1/set":
     if r.Method == "POST" {
       faas.PostTestKV(w,r)
       return
     }
    case "/api/v1/get":
     if r.Method == "GET" {
       faas.GetTestKV(w,r)
       return
     }

    default:
     if r.Method == "GET" {
       fileServer.ServeHTTP(w,r)
       return
     }
  }
  http.NotFound(w, r)
}
*/


func (e *Storage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
  http.NotFound(w, r)
}


/*

  router := httprouter.New()
  router.NotFound = http.FileServer(http.Dir("public")).ServeHTTP
  log.Fatal(http.ListenAndServe(":8000", router))
*/


func main() {


  var cfg_file_name* string = flag.String("cfg", "config.json", "config file")  
  var db_file_name* string = flag.String("db", "data.db", "db file")  
  var cfg_id* int = flag.Int("id", 0, "id")  

  var err error
  runtime.GOMAXPROCS(runtime.NumCPU() - 1)
//  runtime.GOMAXPROCS(1)


// read config
  cfg_file, e := ioutil.ReadFile(*cfg_file_name)
  if e != nil {
    fmt.Printf("File error: %v\n", e)
    os.Exit(1)
  }
  fmt.Printf("%s\n", string(cfg_file))

  flag.Parse()

  err = json.Unmarshal(cfg_file, &cfg)
  if err != nil {
    fmt.Printf("Cfg file json error: %v\n", err)
    os.Exit(1)
  }
  cfg.Id = uint(*cfg_id)
  cfg.Stora_host = cfg.Hosts[cfg.Id*2];
  cfg.Web_host = cfg.Hosts[cfg.Id*2+1];

  fmt.Printf("id: %d\n", *cfg_id)
  fmt.Printf("replica: %d\n", cfg.Replica)
  fmt.Printf("cfg: %s,%s\n", cfg.Stora_host,cfg.Web_host)

  faas.InitDB2(db_file_name,[]string{"users","tests","usersids"})
  defer faas.FinitDB2()

// server setup
  finish := make(chan bool)


  go func() {
    log.Println("http.ListenAndServe Stora_host "+cfg.Stora_host);
    h:=strings.Split(cfg.Stora_host, ":")  
    log.Fatal(http.ListenAndServe(":"+h[1], &Storage{}))
  }()

  go func() {    
    log.Println("http.ListenAndServe Web_host "+cfg.Web_host);    
    h:=strings.Split(cfg.Web_host, ":")
    log.Fatal(faas.ListenAndServe(":"+h[1],"public"))
  }()

  <-finish
}
