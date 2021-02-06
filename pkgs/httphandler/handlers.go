// Package httphandler is used to render webpages.
package httphandler

import (
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"

	"github.com/AJ2O/bytesizelinks/pkgs/api"
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

// HomePageHandler renders the home page.
// e.g. r.HandleFunc("/", HomePageHandler).Methods("GET")
func HomePageHandler(w http.ResponseWriter, r *http.Request) {
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
}

// GenerateLinkHandler renders the page when a Byte-Link is generated.
// e.g. r.HandleFunc("/", GenerateLinkHandler).Methods("POST")
func GenerateLinkHandler(w http.ResponseWriter, r *http.Request) {
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

	link, err := api.GenerateByteLink(sourceLink, customLink)
	if err != nil {
		data.ErrorMessage = true
		data.Message = err.Error()
	} else {
		data.GeneratedMessage = true
		data.Message = link
		log.Println(r.Method, sourceLink, link)
	}

	tmpl.Execute(w, data)
}

// RedirectByteLinkHandler attempts to re-route the request using a provided byte-link.
// e.g. r.HandleFunc("/{byteLink}", HomePageHandler).Methods("GET")
func RedirectByteLinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	byteLink := vars["byteLink"]

	link, err := api.GetOriginalURL(byteLink)
	if err != nil {
		// render error page if the link doesn't work
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
}
