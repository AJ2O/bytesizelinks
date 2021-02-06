package inputvalidation

import (
	"testing"
)

func TestValidateSourceLink(t *testing.T) {
	cases := []struct {
		in, errorMessage string
	}{
		// blank input
		{"", "Please enter a link!"},
		{"   ", "Please enter a link!"},

		// relative paths
		{"/", "Please enter a valid URL!"},
		{"/foo/bar", "Please enter a valid URL!"},

		// IP addresses (unsupported for now)
		{"127.0.0.1", "Please enter a valid URL!"},
		{"127.0.0.1:80", "Please enter a valid URL!"},

		// transport protocol
		{"http", "Please enter a valid URL!"},
		{"http://", "Please enter a valid URL!"},

		// domain resolution
		{"http://www", "Please enter a valid URL!"},
		{"google.com", "Please enter a valid URL!"},
		{"http://google.com", ""},
		{"http://google.com:443", ""},
	}
	for _, c := range cases {
		got := ValidateSourceLink(c.in)
		if got == nil {
			if c.errorMessage != "" {
				t.Errorf("ValidateSourceLink(%q) == nil, want Error(%q)", c.in, c.errorMessage)
			}
		} else {
			if got.Error() != c.errorMessage {
				t.Errorf("ValidateSourceLink(%q) == %q, want Error(%q)", c.in, got.Error(), c.errorMessage)
			}
		}
	}
}

func TestValidateCustomLink(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", "Please enter a link!"},
	}
	for _, c := range cases {
		got := ValidateSourceLink(c.in)
		if got == nil {
			t.Errorf("ValidateSourceLink(%q) == nil, want Error(%q)", c.in, c.want)
		}
		if got.Error() != c.want {
			t.Errorf("ValidateSourceLink(%q) == %q, want Error(%q)", c.in, got.Error(), c.want)
		}
	}
}
