package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

const (
	defaultPort = "21"
)

var allowed = make(map[string]url.URL, 10)
var accessKey string

func get(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		fmt.Fprint(w, "Invalid method")
		return
	}

	query := req.URL.Query()
	token := query.Get("token")

	item, exists := allowed[token]
	if !exists {
		fmt.Fprint(w, "Not found")
		return
	}

	c, err := ftp.Dial(item.Host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "Failed to connect to server")
		return
	}

	password, exists := item.User.Password()
	if !exists {
		password = ""
	}
	err = c.Login(item.User.Username(), password)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "Failed to login")
		return
	}

	res, err := c.Retr(item.Path)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "Failed to open file")
		return
	}

	io.Copy(w, res)

	delete(allowed, token)
	if err := c.Quit(); err != nil {
		log.Print(err)
		return
	}
}

func open(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		fmt.Fprint(w, "Invalid method")
		return
	}

	keyVal := req.FormValue("key")
	if keyVal != accessKey {
		fmt.Fprint(w, "Invalid key")
		return
	}

	u, err := url.Parse(req.FormValue("url"))
	if err != nil {
		log.Fatal(err)
		fmt.Fprint(w, "Invalid file URL")
		return
	}

	key, err := uuid.NewRandom()
	if err != nil {
		log.Fatal(err)
		fmt.Fprint(w, "Internal error")
		return
	}

	strkey := key.String()
	allowed[strkey] = *u

	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["token"] = strkey

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		fmt.Fprint(w, "Internal error")
	}
	w.Write(jsonResp)
	return
}

func main() {
	_ = godotenv.Load(".env")

	key, exists := os.LookupEnv("ACCESS_KEY")
	if !exists {
		panic("ACCESS_KEY env value not found")
	}
	accessKey = key

	http.HandleFunc("/get", get)
	http.HandleFunc("/open", open)
	http.ListenAndServe(":2180", nil)
}
