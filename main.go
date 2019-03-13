package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
)

var stravaOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:3000/callback",
	ClientID:     os.Getenv("STRAVA_CLIENT_ID"),
	ClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
	Scopes: []string{
		"activity:read"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://www.strava.com/oauth/authorize",
		TokenURL: "https://www.strava.com/oauth/token",
	},
}

//StravaUser - Struct for a Strava user
//TODO: Add the rest of the fields for the user
type StravaUser struct {
	ID        int    `json:"id"`        //Strava user Id
	UserName  string `json:"username"`  //Strava user name
	FirstName string `json:"firstname"` // Strava user first name
	LastName  string `json:"lastname"`  //Strava user last name
}

//StravaActivity - Struct for an activity from Strava
//TODO: Add more fields
type StravaActivity struct {
	ID         int     `json:"id"`          //Activity ID
	Name       string  `json:"name"`        //Activity Name
	Distance   float32 `json:"distance"`    //Activity Distance
	MovingTime float32 `json:"moving_time"` //Activity Moving Time - different from Elapsed Time
}

//https://www.strava.com/oauth/authorize?client_id=&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback&response_type=code&scope=activity%3Awrite%2Cread&state=hvilNIqpcZ1uD5mn

func main() {

	//var dir string

	//flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	//flag.Parse()

	router := mux.NewRouter()

	//TODO: Tidy up the code, add templates for the static htmls
	//TODO: Move the handlers to a seperate .go file.
	//TODO: Remove all commented code that is not in use.
	//router.HandleFunc("/", indexHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/callback", callbackHandler)
	router.HandleFunc("/activities", activitiesHandler)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	//http.Handle("/", router)

	http.ListenAndServe(":3000", router)
}

//func indexHandler(w http.ResponseWriter, r *http.Request) {}

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

	fmt.Fprintf(w, "AccessToken: %s: \n", token.AccessToken)

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
	fmt.Fprintf(w, "User: %s(%d)\nName: %s %s\n", user.UserName, user.ID, user.FirstName, user.LastName)

	fmt.Fprintln(w, "<a href='/activities'>Get Activities</a>")
	fmt.Fprintln(w, "<a href='/login'>Log in with Strava</a>")
}

func activitiesHandler(w http.ResponseWriter, r *http.Request) {
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

	//"https://www.strava.com/api/v3/athlete/activities?before=&after=&page=&per_page=" "Authorization: Bearer [[token]]"

	response, err := http.Get("https://www.strava.com/api/v3/athlete/activities?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Error getting user from token %s\n", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	log.Printf("%s\n", contents)

	var activities []StravaActivity
	err = json.Unmarshal([]byte(contents), &activities)
	if err != nil {
		log.Printf("Error unmarshaling Strava Activities %s\n", err.Error())
		return
	}

	fmt.Printf("Activities: %+v\n", activities)
}
