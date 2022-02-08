package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type GitHubContentResponse struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (api *API) fetchGitHubFile(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		req.Header.Add("authorization", "token "+githubToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	githubResp := GitHubContentResponse{}
	err = json.NewDecoder(resp.Body).Decode(&githubResp)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if githubResp.Content == "" {
		return nil, errors.New("missing content")
	}
	if githubResp.Encoding != "base64" {
		return nil, fmt.Errorf("unknown encoding %s", githubResp.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(githubResp.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("decode base64 content: %w", err)
	}
	return decoded, nil
}
