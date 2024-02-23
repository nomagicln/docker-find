package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FindOptions struct {
	// The name of the image to find.
	Name string

	// The maximum number of images to return.
	Page int

	// The number of images to return per page.
	PageSize int
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

type FindFunc func() ([]Image, FindFunc, error)

func nextFindFunc(ctx context.Context, name string, url string) FindFunc {
	return func() ([]Image, FindFunc, error) {
		images, next, err := fetchImages(ctx, name, url)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch images: %w", err)
		}

		return images, nextFindFunc(ctx, name, next), nil
	}
}

func FindImagesFunc(ctx context.Context, opts FindOptions) FindFunc {
	url := fmt.Sprintf(
		"https://hub.docker.com/v2/namespaces/library/repositories/%s/tags?page=%d&page_size=%d",
		opts.Name, opts.Page, opts.PageSize,
	)
	return nextFindFunc(ctx, opts.Name, url)
}

func fetchImages(ctx context.Context, repository, url string) ([]Image, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send request: %w", err)
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
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
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

	return images, findResult.Next, nil
}
