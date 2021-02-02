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

	// home
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/home.html")
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

		// post-request -> generate URL
		if r.Method == http.MethodPost {
			sourceLink := r.FormValue("sourceLink")
			requestLink := r.FormValue("requestLink")

			link, err := generateLink(sourceLink, requestLink)
			if err != nil {
				data.ErrorMessage = true
			} else {
				data.GeneratedMessage = true
				data.Message = link
			}
		}

		tmpl.Execute(w, data)
	})

	// re-routing to other pages
	r.HandleFunc("/{byteLink}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		byteLink := vars["byteLink"]

		link, err := getLink(byteLink)
		if err != nil {
			//
		} else {
			// re-route to website
			log.Println(link)
			http.Redirect(w, r, link, http.StatusSeeOther)
		}
	})

	http.ListenAndServe(":80", r)
}
