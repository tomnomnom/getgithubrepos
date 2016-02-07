package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tomnomnom/linkheader"
)

// repoError is a special error type that includes an exit code
type repoError struct {
	error
	code int
}

// Error values
var errNoUsername = &repoError{errors.New("Usage: getgithubrepos <username>"), 1}
var errBadUsername = &repoError{errors.New("No such username"), 2}
var errRateLimitExceeded = &repoError{errors.New("Rate limit exceeded"), 3}
var errHTTPFail = &repoError{errors.New("HTTP Error"), 4}
var errJSONDecode = &repoError{errors.New("Failed to decode JSON response"), 5}

// handleError takes appropriate action for the provided error
func handleError(err *repoError) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(err.code)
}

func main() {
	flag.Parse()

	user := flag.Arg(0)
	if user == "" {
		handleError(errNoUsername)
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", user)

	r := struct {
		repos []repo
		url   string
	}{
		repos: make([]repo, 0),
		url:   url,
	}

	for r.url != "" {
		fetched, nextURL, err := getRepos(r.url)
		handleError(err)
		r.repos = append(r.repos, fetched...)
		r.url = nextURL
	}

	for _, i := range r.repos {
		fmt.Printf("%s\n", i.SSHUrl)
	}
}

// repo is a struct to unmarshal the JSON response in to
type repo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	SSHUrl string `json:"ssh_url"`
}

// getRepos gets the repositories from a GitHub API URL
// e.g. https://api.github.com/users/tomnomnom/repos
// also returns the URL for the next page of results (if any)
// and any error that occurred
func getRepos(url string) ([]repo, string, *repoError) {

	var repos []repo

	resp, err := http.Get(url)
	if err != nil {
		return repos, "", errHTTPFail
	}
	defer resp.Body.Close()

	// Check for 'expected' errors
	switch resp.StatusCode {
	case http.StatusNotFound:
		return repos, "", errBadUsername
	case http.StatusForbidden:
		return repos, "", errRateLimitExceeded
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return repos, "", errJSONDecode
	}

	// Check for a link to the next page
	next := ""
	if links, exists := resp.Header["Link"]; exists {
		next = getNext(links)
	}

	return repos, next, nil
}

// getNext looks for a rel="next" Link and returns it if it exists
func getNext(headers []string) string {
	links := linkheader.ParseMultiple(headers).FilterByRel("next")
	if len(links) > 0 {
		return links[0].URL
	}
	return ""
}
