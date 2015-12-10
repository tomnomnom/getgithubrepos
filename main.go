package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tomnomnom/linkheader"
)

type repo struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	SSHUrl string `json:"ssh_url"`
}

func main() {
	flag.Parse()

	user := flag.Arg(0)
	if user == "" {
		fmt.Println("Usage: getgithubrepos <username>")
		return
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
		fetched, nextUrl, err := getRepos(r.url)
		if err != nil {
			log.Fatal(err)
			break
		}
		r.repos = append(r.repos, fetched...)
		r.url = nextUrl
	}

	for _, i := range r.repos {
		fmt.Printf("%s\n", i.SSHUrl)
	}
}

// getRepos gets the repositories from a GitHub API URL
// e.g. https://api.github.com/users/tomnomnom/repos
// and also returns the URL for the next page of results (if any)
func getRepos(url string) (repos []repo, next string, err error) {

	resp, err := http.Get(url)
	if err != nil {
		return repos, "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	next = ""
	if links, exists := resp.Header["Link"]; exists {
		next = getNext(links)
	}

	err = json.Unmarshal(body, &repos)
	if err != nil {
		return repos, "", err
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
