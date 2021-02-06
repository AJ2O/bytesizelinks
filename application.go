package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
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

// ErrorPage defines values used for display on the error page
type ErrorPage struct {
	Message string
}

const byteSizeLinksURL = "https://bytesize.link"
const apiURL = "https://api.bytesize.link"

func generateLink(sourceLink string, customLink string) (string, error) {
	// TODO: validate source link
	// 1. Empty?
	if len(sourceLink) == 0 {
		return "", errors.New("Please enter a link!")
	}

	// 2. TODO: Valid Link?

	// 3. Custom link must be alphanumeric
	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(customLink) {
		return "", errors.New("The custom link may only contain numbers or letters!")
	}

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
	// 1. Empty?
	if len(byteLink) == 0 {
		return "", errors.New("Please enter a byte-link!")
	}

	// 2. Must be alphanumeric
	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(byteLink) {
		return "", errors.New("This byte-link is invalid!")
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
			// error page
			log.Println("ERROR", byteLink, link)

			tmpl, tmplErr := template.ParseFiles("html/error.html")
			if tmplErr != nil {
				log.Fatal(tmplErr)
			}
			data := ErrorPage{
				Message: err.Error(),
			}
			tmpl.Execute(w, data)

		} else {
			// re-route to website
			log.Println(r.Method, byteLink, link)
			http.Redirect(w, r, link, http.StatusSeeOther)
		}
	}).Methods("GET")

	// start web server
	http.ListenAndServe(":80", r)
}
