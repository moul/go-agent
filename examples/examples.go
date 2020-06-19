package examples

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIURL is a sample API endpoint used in the examples. It is an actual endpoint
// for the Github v3 API, so may be subject to rate limitations.
var APIURL = `https://api.github.com/orgs/Bearer`

type githubOrg struct {
	Name        string
	Description string
	Blog        string
}

type gogsOrg struct {
	Name        string `json:"full_name"`
	Description string `json:"description"`
	Blog        string `json:"website"`
}

// ShowGithubOrg is a helper method displaying selected parts of a Github "githubOrg repos"
// API response.
func ShowGithubOrg(body []byte) {
	org := githubOrg{}
	json.Unmarshal(body, &org)
	fmt.Println(strings.Join([]string{
		org.Name,
		org.Description,
		org.Blog,
	}, "\n"))
}

// ShowGogsOrg is a helper method displaying selected parts of a Gogs "Org repos"
// API response.
func ShowGogsOrg(body []byte) {
	org := gogsOrg{}
	json.Unmarshal(body, &org)
	fmt.Println(strings.Join([]string{
		org.Name,
		org.Description,
		org.Blog,
	}, "\n"))
}

