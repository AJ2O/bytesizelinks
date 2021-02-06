package main

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/AJ2O/bytesizelinks/pkgs/httphandler"
)

func main() {
	// router creation
	r := mux.NewRouter()

	// serve static resources (ex. CSS)
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// home page
	r.HandleFunc("/", httphandler.HomePageHandler).Methods("GET")

	// generate byte-link
	r.HandleFunc("/", httphandler.GenerateLinkHandler).Methods("POST")

	// re-direct with byte-link
	r.HandleFunc("/{byteLink}", httphandler.RedirectByteLinkHandler).Methods("GET")

	// start web server
	http.ListenAndServe(":80", r)
}
