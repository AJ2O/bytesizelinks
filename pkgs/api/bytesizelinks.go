// Package api handles invocations to backend APIs.
package api

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/AJ2O/bytesizelinks/pkgs/inputvalidation"
)

const apiURL = "https://api.bytesize.link"

// GenerateByteLink creates a byte-link of the given source URL.
// If a custom link is specified, it tries to use the custom link as the byte-link.
func GenerateByteLink(sourceURL string, customLink string) (string, error) {
	// validate source link
	err := inputvalidation.ValidateSourceLink(sourceURL)
	if err != nil {
		return "", err
	}

	// validate custom link
	err = inputvalidation.ValidateCustomLink(customLink)
	if err != nil {
		return "", err
	}
	// Invoke API
	queryParams := "?sourceLink=" + sourceURL + "&customByteLink=" + customLink
	response, err := http.Post(
		apiURL+"/"+queryParams,
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

// GetOriginalURL returns the original URL that the given byte-link is mapped to.
func GetOriginalURL(byteLink string) (string, error) {
	// validate byte-link
	err := inputvalidation.ValidateByteLink(byteLink)
	if err != nil {
		return "", err
	}

	// Invoke API
	queryParams := "?byteLink=" + byteLink
	response, err := http.Get(apiURL + "/" + queryParams)
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
