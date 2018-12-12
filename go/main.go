package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	pusher "github.com/pusher/pusher-http-go"
)

type User struct {
	Email    string
	Password string
}

var (
	validUsers = map[string]User{
		"admin": User{
			Email:    "yo@lanre.wtf",
			Password: "admin",
		},
		"lanre": User{

			Email:    "yo@lanre.wtf",
			Password: "lanre",
		},
	}
)

func main() {

	port := flag.Int("http.port", 1400, "Port to run HTTP service on")

	flag.Parse()

	appID := os.Getenv("PUSHER_APP_ID")
	appKey := os.Getenv("PUSHER_APP_KEY")
	appSecret := os.Getenv("PUSHER_APP_SECRET")
	appCluster := os.Getenv("PUSHER_APP_CLUSTER")
	appIsSecure := os.Getenv("PUSHER_APP_SECURE")

	var isSecure bool
	if appIsSecure == "1" {
		isSecure = true
	}

	client := &pusher.Client{
		AppId:   appID,
		Key:     appKey,
		Secret:  appSecret,
		Cluster: appCluster,
		Secure:  isSecure,
	}

	mux := http.NewServeMux()

	mux.Handle("/login", http.HandlerFunc(login(client)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}

func encode(w io.Writer, v interface{}) {
	json.NewEncoder(w).Encode(v)
}

func login(client *pusher.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var request struct {
			UserName string `json:"userName"`
			Password string `json:"password"`
		}

		type response struct {
			Message string `json:"message"`
			Success bool   `json:"success"`
		}

		if r.URL.Path != "/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encode(w, response{"Invalid request body", false})
			return
		}

		user, ok := validUsers[request.UserName]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			encode(w, response{"User not found", false})
			return
		}

		if user.Password != request.Password {
			w.WriteHeader(http.StatusBadRequest)
			encode(w, response{"Password does not match", false})
			return
		}

		w.WriteHeader(http.StatusOK)
		encode(w, response{"Login successful", true})

		_, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
			return
		}

		client.Trigger("auth", "login", &struct {
			IP    string `json:"ip"`
			User  string `json:"user"`
			Email string `json:"email"`
		}{
			User:  request.UserName,
			IP:    r.RemoteAddr,
			Email: user.Email,
		})
	}
}
