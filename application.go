package main

import (
	//"bytes"
	//"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

// HomePage defines values used for display on the home page
type HomePage struct {
	WebsiteURL       string
	GeneratedURL     string
	Message          string
	GeneratedMessage bool
	ErrorMessage     bool
}

const byteSizeLinksURL = "https://bytesize.link"
const apiURL = "https://api.bytesize.link"

func generateLink(sourceLink string, requestLink string) (string, error) {
	// TODO: validate source link
	// 1. Empty?
	// 2. Valid Link?

	// Invoke API
	response, err := http.Post(apiURL+"/?sourceLink="+sourceLink, "text/plain", nil)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(response.Body)
	stringData := string(data)
	stringData = stringData[1 : len(stringData)-1]

	return stringData, nil
}
func getLink(byteLink string) (string, error) {
	// TODO: validate input link
	// 1. Empty?

	// Invoke API
	response, err := http.Get(apiURL + "/?byteLink=" + byteLink)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(response.Body)
	stringData := string(data)
	stringData = stringData[1 : len(stringData)-1]

	return stringData, nil
}

func main() {
	// router creation
	r := mux.NewRouter()

	// serve static resources (ex. CSS)
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// GET: home page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("html/home.html")
		if err != nil {
			log.Fatal(err)
		}

		data := HomePage{
			WebsiteURL:       byteSizeLinksURL,
			GeneratedURL:     "",
			Message:          "",
			ErrorMessage:     false,
			GeneratedMessage: false,
		}

		tmpl.Execute(w, data)
	}).Methods("GET")

	// POST: process url request
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("html/home.html")
		if err != nil {
			log.Fatal(err)
		}

		data := HomePage{
			WebsiteURL:       byteSizeLinksURL,
			GeneratedURL:     "",
			Message:          "",
			ErrorMessage:     false,
			GeneratedMessage: false,
		}
		sourceLink := r.FormValue("sourceLink")
		requestLink := r.FormValue("requestLink")

		link, err := generateLink(sourceLink, requestLink)
		if err != nil {
			data.ErrorMessage = true
		} else {
			data.GeneratedMessage = true
			data.Message = link
			log.Println(r.Method, sourceLink, link)
		}

		tmpl.Execute(w, data)
	}).Methods("POST")

	// GET: re-routing to other pages
	r.HandleFunc("/{byteLink}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		byteLink := vars["byteLink"]

		link, err := getLink(byteLink)
		if err != nil {
			//
		} else {
			// re-route to website
			log.Println(r.Method, byteLink, link)
			http.Redirect(w, r, link, http.StatusSeeOther)
		}
	}).Methods("GET")

	http.ListenAndServe(":80", r)
}
