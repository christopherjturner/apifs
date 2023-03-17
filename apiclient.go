package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Team struct {
	Name  string // json:"name"
	Repos int    // json:"repos"
}

var client http.Client = http.Client{}

func GetTeams() ([]Team, error) {
	resp, err := client.Get("https://catalogue.tax.service.gov.uk/api/v2/teams")
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	teams := []Team{}
	err = json.Unmarshal(body, &teams)

	return teams, err
}

type Named struct {
	Name string // json:"name"
}

func GetRepos(repoType string) ([]Named, error) {

	url := fmt.Sprintf("https://catalogue.tax.service.gov.uk/api/v2/repositories?archived=false&repoType=%s", repoType)
	resp, err := client.Get(url)

	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	repos := []Named{}
	err = json.Unmarshal(body, &repos)

	return repos, err
}

func GetRepo(name string) (string, error) {
	url := fmt.Sprintf("https://catalogue.tax.service.gov.uk/api/v2/repositories/%s", name)
	resp, err := client.Get(url)

	if err != nil {
		return "", err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
