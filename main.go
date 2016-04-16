// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"log"
	"net/url"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	addr = flag.String("addr", "127.0.0.1:8080", "http service address")
	username = flag.String("username", "", "chat username")
	register = flag.Bool("register", false, "set to register")
)

const not_username string = ""
const error_not_username = "no username"

type registerData struct {
  Username string
  Password string
}

func get_auth(username string, password string) http.Header {
	req,_ := http.NewRequest("GET", "/", nil)
	req.SetBasicAuth(username,password)
	return req.Header
}

func recive_message(c *websocket.Conn, done chan struct{}){
	defer c.Close()
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}
}

func scan_message(text chan string){
	defer close(text)
	for {
	  reader := bufio.NewReader(os.Stdin)
	  in, err := reader.ReadString('\n')
	  if err == nil {
	    text <- in
    }
  }
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

func scan_password() string {
	log.Printf("type your password: ")
  pass,err := gopass.GetPasswd()
  if err == nil {
    return string(pass)
  }
	return ""
}

func validate_username(username string) bool {
	if strings.Compare(username, not_username) != 0 {
	  return true
	}	else {
		return false
	}
}

func generate_credential(username *string) (http.Header, error) {
	err := errors.New(error_not_username)
	auth_header := new(http.Header)
	if validate_username(*username){
		err = nil
		pass := scan_password()
		*auth_header = get_auth(*username, pass)
	}
	return *auth_header, err
}

func print_error_and_exit(err error, code int){
	log.Println(err)
	os.Exit(code)
}

func get_websocket_connection(auth_header http.Header)(*websocket.Conn, error){
	u := url.URL{Scheme: "ws", Host: *addr, Path: "wsc"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), auth_header)
	return c, err
}

func start_scan_routine(c *websocket.Conn, done chan struct{}, text chan string){
	go recive_message(c, done)
	go scan_message(text)
}

func chat_routine(c *websocket.Conn, interrupt chan os.Signal){
	done := make(chan struct{})
	text := make(chan string)
	start_scan_routine(c, done, text)

	for {
		select {
		case t:= <-text:
			err := c.WriteMessage(websocket.TextMessage, []byte(t))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

func main() {
	//prepare initial
	flag.Parse()
	log.SetFlags(0)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if *register {
		name := scan_username()
		pass := scan_password()

		url := fmt.Sprintf("http://%s/cregister", *addr)
		json_value := registerData{Username: name, Password: pass}

		b := new(bytes.Buffer)
    json.NewEncoder(b).Encode(json_value)
    res, _ := http.Post(url, "application/json; charset=utf-8", b)

    log.Println("response Status:", res.Status)
    log.Println("response Headers:", res.Header)
    body, _ := ioutil.ReadAll(res.Body)
    log.Println("response Body:", string(body))

	} else {
		//generate_credential
		//auto exit if error
		auth_header,err := generate_credential(username)
		if err != nil { print_error_and_exit(err, 101) }

		//get websocket connection
		//auto exit if error
		c, err := get_websocket_connection(auth_header)
		if err != nil { print_error_and_exit(err, 102) }
		defer c.Close()

		chat_routine(c, interrupt)
	}
}
