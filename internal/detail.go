package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Detail struct {
	User              string    `json:"user"`
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	RepositoryType    string    `json:"repository_type"`
	Status            int       `json:"status"`
	StatusDescription string    `json:"status_description"`
	Description       string    `json:"description"`
	IsPrivate         bool      `json:"is_private"`
	IsAutomated       bool      `json:"is_automated"`
	StarCount         int       `json:"star_count"`
	PullCount         int64     `json:"pull_count"`
	LastUpdated       time.Time `json:"last_updated"`
	DateRegistered    time.Time `json:"date_registered"`
	CollaboratorCount int       `json:"collaborator_count"`
	Affiliation       any       `json:"affiliation"`
	HubUser           string    `json:"hub_user"`
	HasStarred        bool      `json:"has_starred"`
	FullDescription   string    `json:"full_description"`
	Permissions       struct {
		Read  bool `json:"read"`
		Write bool `json:"write"`
		Admin bool `json:"admin"`
	} `json:"permissions"`
	MediaTypes   []string `json:"media_types"`
	ContentTypes []string `json:"content_types"`
}

func GetDetail(name string) (Detail, error) {
	return fetchDetail(name)
}

func fetchDetail(name string) (Detail, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/library/repositories/%s", name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Detail{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Detail{}, fmt.Errorf("failed to send request: %w", err)
	}

	var detail Detail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return Detail{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return detail, nil
}
