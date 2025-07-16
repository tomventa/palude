package ollama

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/tomventa/palude/internal/config"
	"github.com/tomventa/palude/internal/types"
)

// Client represents an Ollama API client
type Client struct {
	config *config.Config
}

// New creates a new Ollama client
func New(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// Query sends a prompt to Ollama and returns the response
func (c *Client) Query(prompt string) (string, error) {
	reqBody := types.OllamaRequest{
		Model:       c.config.Model,
		Prompt:      prompt,
		Stream:      false,
		Temperature: 0.95,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.config.OllamaURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp types.OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
