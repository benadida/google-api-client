package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"bytes"
	"encoding/gob"
	"encoding/base64"
	"net/http"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

func main() {
	google_client_id := os.Getenv("GOOGLE_CLIENT_ID")
	google_client_secret := os.Getenv("GOOGLE_CLIENT_SECRET")

	var config = &oauth2.Config{
		ClientID:     google_client_id, 
		ClientSecret: google_client_secret, 
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveScope},
	}

	ctx := context.Background()

	tokenFromWeb(ctx, config)
}

func tokenFromWeb(ctx context.Context, config *oauth2.Config) {
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())

	config.RedirectURL = "http://localhost:3000/after"

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("====> Got: %s", req.URL.Path)
		
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		
		if req.URL.Path == "/" {
			authURL := config.AuthCodeURL(randState)
			http.Redirect(rw, req, authURL, http.StatusFound)
		}
		
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: %s / %s ", req.FormValue("state"), randState)
			http.Error(rw, "", 500)
			return
		}
		
		if code := req.FormValue("code"); code != "" {
			log.Printf("sending code")
			
			token, err := config.Exchange(ctx, code)

			if err == nil {
				// encode this thing
				var tokenBytes bytes.Buffer 
				enc := gob.NewEncoder(&tokenBytes)
				enc.Encode(token)
				tokenString := base64.StdEncoding.EncodeToString(tokenBytes.Bytes())
				log.Printf(tokenString)
			}
			
			log.Printf("sent code")
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	})
	
	http.ListenAndServe(":3000", nil)
}
