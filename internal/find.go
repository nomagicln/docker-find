package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type FindOptions struct {
	// The name of the image to find.
	Name string

	// The tag of the image.
	Tag string

	// The number of images to return.
	Limit int

	// The date the image was created after.
	After time.Time
}

// Image represents a Docker image.
type Image struct {
	// The repository of the image.
	Repository string

	// The tag of the image.
	Tag string

	// The size of the image.
	Size int

	// The date the image was created.
	Created time.Time

	// The digest of the image.
	Digest string
}

// FindImages returns a list of images that match the given options.
func FindImages(opts FindOptions) ([]Image, error) {
	// If the name is empty, return an error.
	if opts.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// If the limit is less than 1, return an empty list.
	if opts.Limit < 1 {
		return nil, nil
	}

	// Fetch the images from the Docker Hub API.
	images, err := fetchImages(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch images: %w", err)
	}

	// Return the list of images.
	return filterImages(images, opts), nil
}

func fetchImages(repository string) ([]Image, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/namespaces/library/repositories/%s/tags", repository)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var findResult = struct {
		Count    int    `json:"count"`
		Next     string `json:"next"`
		Previous any    `json:"previous"`
		Results  []struct {
			Creator int `json:"creator"`
			ID      int `json:"id"`
			Images  []struct {
				Architecture string    `json:"architecture"`
				Features     string    `json:"features"`
				Variant      any       `json:"variant"`
				Digest       string    `json:"digest"`
				Os           string    `json:"os"`
				OsFeatures   string    `json:"os_features"`
				OsVersion    any       `json:"os_version"`
				Size         int       `json:"size"`
				Status       string    `json:"status"`
				LastPulled   time.Time `json:"last_pulled"`
				LastPushed   time.Time `json:"last_pushed"`
			} `json:"images"`
			LastUpdated         time.Time `json:"last_updated"`
			LastUpdater         int       `json:"last_updater"`
			LastUpdaterUsername string    `json:"last_updater_username"`
			Name                string    `json:"name"`
			Repository          int       `json:"repository"`
			FullSize            int       `json:"full_size"`
			V2                  bool      `json:"v2"`
			TagStatus           string    `json:"tag_status"`
			TagLastPulled       time.Time `json:"tag_last_pulled"`
			TagLastPushed       time.Time `json:"tag_last_pushed"`
			MediaType           string    `json:"media_type"`
			ContentType         string    `json:"content_type"`
			Digest              string    `json:"digest"`
		} `json:"results"`
	}{}

	// Decode the response body into the findResult variable.
	if err := json.NewDecoder(resp.Body).Decode(&findResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Create a slice of Image objects to return.
	var images []Image

	// Iterate over the results and create an Image object for each one.
	for _, result := range findResult.Results {
		images = append(images, Image{
			Repository: repository,
			Tag:        result.Name,
			Size:       result.FullSize,
			Created:    result.LastUpdated,
			Digest:     result.Digest,
		})
	}

	return images, nil
}

func filterImages(images []Image, opts FindOptions) []Image {
	// Create a slice to hold the filtered images.
	var filtered []Image

	// Iterate over the images and add the ones that match the options to the filtered slice.
	for _, img := range images {
		if opts.Tag != "" && strings.Contains(img.Tag, opts.Tag) {
			continue
		}

		if !img.Created.After(opts.After) {
			continue
		}

		filtered = append(filtered, img)
	}

	return filtered[:min(opts.Limit, len(filtered))]
}
