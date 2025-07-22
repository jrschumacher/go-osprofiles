package profiles

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var profileNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-_]?[a-z0-9])*$`)
var reverseDNSRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)

func validateProfileName(n string) error {
	// check profile name is valid [a-zA-Z0-9_-]
	if n == "" {
		return fmt.Errorf("%w, profile name: ''", ErrMissingProfileName)
	}
	// check profile name is valid [a-zA-Z0-9_-]
	if !profileNameRegex.MatchString(n) {
		return errors.New("profile name must be alphanumeric with dashes or underscores (e.g. my-profile-name)")
	}
	return nil
}

func validateReverseDNS(domain string) error {
	if domain == "" {
		return errors.New("reverse DNS identifier cannot be empty")
	}
	// Reverse DNS must contain at least one dot
	if !strings.Contains(domain, ".") {
		return errors.New("reverse DNS identifier must contain at least one dot (e.g. com.example.myapp)")
	}
	if !reverseDNSRegex.MatchString(domain) {
		return errors.New("reverse DNS identifier must be in format com.company.app (e.g. com.example.myapp)")
	}
	return nil
}
