package api

import (
	"testing"
)

func TestGenerateByteLink(t *testing.T) {
	cases := []struct {
		sourceLink, customLink string
	}{
		// invalid links
		{"", ""},
		{"  ", ""},
		{"https://google.com", "Adsa$#4"},
	}
	for _, c := range cases {
		byteLink, err := GenerateByteLink(c.sourceLink, c.customLink)
		if err == nil {
			t.Errorf("GenerateByteLink(%q, %q) == (%q, nil), want (\"\", Error())",
				c.sourceLink, c.customLink, byteLink)
		}
	}
}

func TestGetOriginalURL(t *testing.T) {
	cases := []struct {
		byteLink string
	}{
		// invalid link
		{""},
		{"%asd8jn"},
	}
	for _, c := range cases {
		originalURL, err := GetOriginalURL(c.byteLink)
		if err == nil {
			t.Errorf("GetOriginalURL(%q) == (%q, nil), want (\"\", Error())", c.byteLink, originalURL)
		}
	}
}
