package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	notionAPIURL     = "https://api.notion.com/v1"
	notionAuthURL    = "https://api.notion.com/v1/oauth/authorize"
	notionTokenURL   = "https://api.notion.com/v1/oauth/token"
	notionAPIVersion = "2022-06-28"
)

// Config holds Notion OAuth configuration
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Client is the Notion API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Notion API client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type"`
	BotID         string `json:"bot_id"`
	WorkspaceName string `json:"workspace_name"`
	WorkspaceIcon string `json:"workspace_icon"`
	WorkspaceID   string `json:"workspace_id"`
	Owner         struct {
		Type string `json:"type"`
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"owner"`
	DuplicatedTemplateID string `json:"duplicated_template_id"`
}

// GetAuthURL generates the OAuth authorization URL
func (c *Client) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", c.config.ClientID)
	params.Set("redirect_uri", c.config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("owner", "user")
	params.Set("state", state)

	return fmt.Sprintf("%s?%s", notionAuthURL, params.Encode())
}

// ExchangeCode exchanges an authorization code for an access token
func (c *Client) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	data := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": c.config.RedirectURI,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", notionTokenURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// Database represents a Notion database
type Database struct {
	ID    string `json:"id"`
	Title []struct {
		PlainText string `json:"plain_text"`
	} `json:"title"`
}

// SearchDatabases searches for databases the integration has access to
func (c *Client) SearchDatabases(ctx context.Context, accessToken string) ([]Database, error) {
	data := map[string]interface{}{
		"filter": map[string]string{
			"value":    "database",
			"property": "object",
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", notionAPIURL+"/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search databases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(body))
	}

	var result struct {
		Results []Database `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Results, nil
}

// PageResponse represents a created Notion page
type PageResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// CreatePage creates a new page in a database
func (c *Client) CreatePage(ctx context.Context, accessToken string, databaseID string, properties map[string]interface{}, children []interface{}) (*PageResponse, error) {
	data := map[string]interface{}{
		"parent": map[string]string{
			"database_id": databaseID,
		},
		"properties": properties,
	}

	if len(children) > 0 {
		data["children"] = children
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", notionAPIURL+"/pages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("page creation failed: %s", string(body))
	}

	var pageResp PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pageResp, nil
}

// CreateDatabase creates a new database in the workspace
func (c *Client) CreateDatabase(ctx context.Context, accessToken string, parentPageID string, title string) (*Database, error) {
	data := map[string]interface{}{
		"parent": map[string]string{
			"type":    "page_id",
			"page_id": parentPageID,
		},
		"title": []map[string]interface{}{
			{
				"type": "text",
				"text": map[string]string{
					"content": title,
				},
			},
		},
		"properties": map[string]interface{}{
			"Name": map[string]interface{}{
				"title": map[string]interface{}{},
			},
			"Category": map[string]interface{}{
				"select": map[string]interface{}{},
			},
			"Cuisine": map[string]interface{}{
				"rich_text": map[string]interface{}{},
			},
			"Prep Time": map[string]interface{}{
				"number": map[string]interface{}{},
			},
			"Cook Time": map[string]interface{}{
				"number": map[string]interface{}{},
			},
			"Servings": map[string]interface{}{
				"number": map[string]interface{}{},
			},
			"Source URL": map[string]interface{}{
				"url": map[string]interface{}{},
			},
			"Tags": map[string]interface{}{
				"multi_select": map[string]interface{}{},
			},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", notionAPIURL+"/databases", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("database creation failed: %s", string(body))
	}

	var dbResp Database
	if err := json.NewDecoder(resp.Body).Decode(&dbResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dbResp, nil
}
