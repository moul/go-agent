package examples

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// APIURL is a sample API endpoint used in the examples. It is an actual endpoint
// for the Github v3 API, so may be subject to rate limitations.
const APIURL = `https://api.github.com/orgs/Bearer`


type org struct {
	Name        string
	Description string
	Blog        string
}

// ShowOrg is a helper method displaying selected parts of a Github "org repos"
// API response.
func ShowOrg(body []byte) {
	org := org{}
	json.Unmarshal(body, &org)
	fmt.Println(strings.Join([]string{
		org.Name,
		org.Description,
		org.Blog,
	}, "\n"))
}

// MustParse builds a URL instance from a known-good URL string, panicking it
// the URL string is not well-formed.
func MustParse(rawURL string) *url.URL {
	maybeURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return maybeURL
}
