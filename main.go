package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	

	"golang.org/x/oauth2"

	"github.com/gorilla/mux"
	"github.com/dchest/uniuri"
)

var stravaOauthConfig = &oauth2.Config{
	RedirectURL: "http://localhost:3000/callback",
	ClientID: os.Getenv("STRAVA_CLIENT_ID"),
	ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
	Scopes: []string{
	  "activity:write,read"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://www.strava.com/oauth/authorize",
		TokenURL: "https://www.strava.com/oauth/token",
	},
  }

  type StravaUser struct {
	UserName string `json:"username"`
	FirstName string `json:"firstname"`
	LastName string `json:"lastname"`
  }

//https://www.strava.com/oauth/authorize?client_id=&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback&response_type=code&scope=activity%3Awrite%2Cread&state=hvilNIqpcZ1uD5mn

func main() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/callback", callbackHandler)

	http.ListenAndServe(":3000", router)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, os.Getenv("STRAVA_CLIENT_ID"))
	fmt.Fprintln(w, "<a href='/login'>Log in with Strava</a>")
  }

  func loginHandler(w http.ResponseWriter, r *http.Request) {
	oauthStateString := uniuri.New()
	url := stravaOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
  }

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := stravaOauthConfig.Exchange(oauth2.NoContext, code)
	
	if err != nil {
		fmt.Fprintf(w, "Code exchange failed with error %s\n", err.Error())
		return
	}

	if !token.Valid() {
		fmt.Fprintln(w, "Retreived invalid token")
		return
	}

	fmt.Fprintln(w, token.AccessToken)

	response, err := http.Get("https://www.strava.com/api/v3/athlete?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error getting user from token %s\n", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	log.Printf("%s\n", contents)

	var user *StravaUser
	err = json.Unmarshal(contents, &user)
	if err != nil {
		log.Printf("Error unmarshaling Google user %s\n", err.Error())
		return
	}

	//fmt.Fprintf(w, "Email: %s\nName: %s\nImage link: %s\n", user.Email, user.Name, user.Picture)
	fmt.Fprintf(w, "User: %s\nName: %s\n", user.UserName, user.FirstName)
}