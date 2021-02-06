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

		// nonsense
		{"amioasd#2", "Please enter a valid URL!"},

		// relative paths
		{"/", "Please enter a valid URL!"},
		{"/foo/bar", "Please enter a valid URL!"},

		// IP addresses (unsupported for now)
		{"127.0.0.1", "Please enter a valid URL!"},
		{"127.0.0.1:80", "Please enter a valid URL!"},

		// transport protocol
		{"http", "Please enter a valid URL!"},
		{"http:", "Please enter a valid URL!"},
		{"http://", "Please enter a valid URL!"},

		// domain resolution
		{"http://www", "Please enter a valid URL!"},
		{"google.com", "Please enter a valid URL!"},
		{"http://google.com", ""},
		{"https://google.com", ""},
		{"https://google.com:443", ""},

		// invalid mixture
		{"http://google.co m", "Please enter a valid URL!"},
	}
	for _, c := range cases {
		got := ValidateSourceLink(c.in)
		if got == nil {
			if c.errorMessage != "" {
				t.Errorf("ValidateSourceLink(%q) == nil, want Error(%q)", c.in, c.errorMessage)
			}
		} else if got.Error() != c.errorMessage {
			t.Errorf("ValidateSourceLink(%q) == %q, want Error(%q)", c.in, got.Error(), c.errorMessage)
		}
	}
}

func TestValidateCustomLink(t *testing.T) {
	cases := []struct {
		in, errorMessage string
	}{
		// blank input
		{"", "The custom link must not be empty!"},
		{"   ", "The custom link must not be empty!"},

		// invalid characters
		{"@", "The custom link may only contain numbers or letters!"},
		{"()", "The custom link may only contain numbers or letters!"},
		{"*", "The custom link may only contain numbers or letters!"},

		// lowercase
		{"a", ""},
		{"abcdefgh", ""},

		// uppercase
		{"Z", ""},
		{"ZYXWVUTS", ""},

		// numbers
		{"0", ""},
		{"98765432", ""},

		// mixture
		{"Ab34Ef78", ""},

		// invalid mixture
		{"ab2ds$a8", "The custom link may only contain numbers or letters!"},
	}
	for _, c := range cases {
		got := ValidateCustomLink(c.in)
		if got == nil {
			if c.errorMessage != "" {
				t.Errorf("ValidateCustomLink(%q) == nil, want Error(%q)", c.in, c.errorMessage)
			}
		} else if got.Error() != c.errorMessage {
			t.Errorf("ValidateCustomLink(%q) == %q, want Error(%q)", c.in, got.Error(), c.errorMessage)
		}
	}
}

func TestValidateByteLink(t *testing.T) {
	cases := []struct {
		in, errorMessage string
	}{
		// blank input
		{"", "Please enter a byte-link!"},
		{"   ", "Please enter a byte-link!"},

		// invalid characters
		{"@", "This byte-link is invalid!"},
		{"()", "This byte-link is invalid!"},
		{"*", "This byte-link is invalid!"},

		// lowercase
		{"a", ""},
		{"abcdefgh", ""},

		// uppercase
		{"Z", ""},
		{"ZYXWVUTS", ""},

		// numbers
		{"0", ""},
		{"98765432", ""},

		// mixture
		{"Ab34Ef78", ""},

		// invalid mixture
		{"ab2ds$a8", "This byte-link is invalid!"},
	}
	for _, c := range cases {
		got := ValidateByteLink(c.in)
		if got == nil {
			if c.errorMessage != "" {
				t.Errorf("ValidateByteLink(%q) == nil, want Error(%q)", c.in, c.errorMessage)
			}
		} else if got.Error() != c.errorMessage {
			t.Errorf("ValidateByteLink(%q) == %q, want Error(%q)", c.in, got.Error(), c.errorMessage)
		}
	}
}
