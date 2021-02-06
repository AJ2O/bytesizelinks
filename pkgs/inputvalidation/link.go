// Package inputvalidation implements additional functions to validate input given to the web client.
package inputvalidation

import (
	"errors"
	"regexp"
)

// ValidateSourceLink checks if the given source link is of proper format.
func ValidateSourceLink(sourceLink string) error {
	// source link must not be empty
	if len(sourceLink) == 0 {
		return errors.New("Please enter a link!")
	}

	// source link must look like a valid URL

	return nil
}

// ValidateCustomLink checks if the given custom link is of proper format.
func ValidateCustomLink(customLink string) error {
	// custom link must be alphanumeric
	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(customLink) {
		return errors.New("The custom link may only contain numbers or letters!")
	}

	return nil
}

// ValidateByteLink checks if the given byte-link is of proper format.
func ValidateByteLink(byteLink string) error {
	// byte-link must not be empty
	if len(byteLink) == 0 {
		return errors.New("Please enter a byte-link!")
	}

	// byte-link must be alphanumeric
	if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(byteLink) {
		return errors.New("This byte-link is invalid!")
	}

	return nil
}
