package main

import(
  "bytes"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
)

type registerRespond struct {
  Code int
  Message string
}

type registerData struct {
  Username string
  Password string
}

func scan_username() string {
	log.Printf("type your username: ")
	var username string
  _, err := fmt.Scanf("%s", &username)
  if err == nil {
    return string(username)
  }
	return ""
}

func registering(addr string) {
  name := scan_username()
  pass := scan_password()

  url := fmt.Sprintf("http://%s/cregister", addr)
  json_value := registerData{Username: name, Password: pass}

  b := new(bytes.Buffer)
  json.NewEncoder(b).Encode(json_value)
  resp, err := http.Post(url, "application/json; charset=utf-8", b)
  if err != nil {
    log.Fatal("Can't Connect to Server")
  }

  var r registerRespond
  err = json.NewDecoder(resp.Body).Decode(&r)
  if err != nil {
    log.Fatal("error 301")
  }

  log.Println(r.Message)
}
