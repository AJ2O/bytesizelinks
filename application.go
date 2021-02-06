package main

import (
	//"bytes"
	//"encoding/json"

	"errors"
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

func generateLink(sourceLink string, customLink string) (string, error) {
	// TODO: validate source link
	// 1. Empty?
	if len(sourceLink) == 0 {
		return "", errors.New("Please enter a link!")
	}

	// 2. Valid Link?

	// Invoke API
	response, err := http.Post(
		apiURL+"/?sourceLink="+sourceLink+"&customByteLink="+customLink,
		"text/plain",
		nil)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(response.Body)
	stringData := string(data)

	// handle potential errors
	if response.StatusCode != 200 {
		log.Println(response.StatusCode, stringData)
		return "", errors.New(stringData)
	}

	return stringData, nil
}
func getLink(byteLink string) (string, error) {
	// TODO: validate input link
	// 1. Empty?
	if len(byteLink) == 0 {
		return "", errors.New("Please enter a byte-link!")
	}

	// Invoke API
	response, err := http.Get(apiURL + "/?byteLink=" + byteLink)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(response.Body)
	stringData := string(data)

	// handle potential errors
	if response.StatusCode != 200 {
		log.Println(response.StatusCode, stringData)
		return "", errors.New(stringData)
	}

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

		errorMessage := r.FormValue("redirectErrorMessage")
		if len(errorMessage) != 0 {
			data.Message = errorMessage
			data.ErrorMessage = true
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
		customLink := r.FormValue("customLink")

		link, err := generateLink(sourceLink, customLink)
		if err != nil {
			data.ErrorMessage = true
			data.Message = err.Error()
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
			// re-direct home with error
			r.Form.Set("redirectErrorMessage", err.Error())
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			// re-route to website
			log.Println(r.Method, byteLink, link)
			http.Redirect(w, r, link, http.StatusSeeOther)
		}
	}).Methods("GET")

	http.ListenAndServe(":80", r)
}
