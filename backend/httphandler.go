package backend

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTTP request structures
type RequestAuth struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type KeyValue struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type RequestData struct {
	Method   string       `json:"method"`
	URL      string       `json:"url"`
	Params   []KeyValue   `json:"params"`
	Headers  []KeyValue   `json:"headers"`
	Body     string       `json:"body"`
	BodyType string       `json:"bodyType"`
	Auth     *RequestAuth `json:"auth"`
}

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	StatusText string            `json:"statusText"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// Workspace management structures
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type WorkspaceData struct {
	Workspaces []Workspace `json:"workspaces"`
}

// Request collection structures
type CollectionRequest struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	CollectionID string      `json:"collectionId"`
	Request      RequestData `json:"request"`
}

type Collection struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	WorkspaceID string              `json:"workspaceId"`
	Requests    []CollectionRequest `json:"requests"`
	CreatedAt   time.Time           `json:"createdAt"`
}

type CollectionData struct {
	Collections []Collection `json:"collections"`
}

// Environment management structures
type Environment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Variables   map[string]string `json:"variables"`
	WorkspaceID string            `json:"workspaceId"`
}

type EnvironmentData struct {
	Environments []Environment `json:"environments"`
}

// Request history structures
type HistoryItem struct {
	ID          string        `json:"id"`
	Request     RequestData   `json:"request"`
	Response    *ResponseData `json:"response"`
	Timestamp   time.Time     `json:"timestamp"`
	WorkspaceID string        `json:"workspaceId"`
}

type HistoryData struct {
	Items []HistoryItem `json:"items"`
}

// Application settings structure
type Settings struct {
	UIScale              int    `json:"uiScale"`
	Theme                string `json:"theme"`
	LayoutMode           string `json:"layoutMode"`
	AutoSaveHistory      bool   `json:"autoSaveHistory"`
	MaxHistoryItems      int    `json:"maxHistoryItems"`
	DefaultTimeout       int    `json:"defaultTimeout"`
	PrettyPrintByDefault bool   `json:"prettyPrintByDefault"`
}

// HTTPHandler manages HTTP-related functionality
type HTTPHandler struct {
	app     AppInterface
	dataDir string
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(app AppInterface, dataDir string) *HTTPHandler {
	return &HTTPHandler{
		app:     app,
		dataDir: dataDir,
	}
}

// SendRequest sends an HTTP request
func (h *HTTPHandler) SendRequest(req RequestData) (*ResponseData, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for _, header := range req.Headers {
		if header.Enabled && header.Key != "" {
			httpReq.Header.Set(header.Key, header.Value)
		}
	}

	if req.Auth != nil {
		switch req.Auth.Type {
		case "basic":
			auth := base64.StdEncoding.EncodeToString([]byte(req.Auth.Username + ":" + req.Auth.Password))
			httpReq.Header.Set("Authorization", "Basic "+auth)
		case "bearer":
			httpReq.Header.Set("Authorization", "Bearer "+req.Auth.Token)
		case "api-key":
			httpReq.Header.Set(req.Auth.Key, req.Auth.Value)
		}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &ResponseData{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       string(bodyBytes),
	}, nil
}

// SaveWorkspaces saves workspaces to file
func (h *HTTPHandler) SaveWorkspaces(workspaces []Workspace) error {
	data := WorkspaceData{Workspaces: workspaces}
	return h.saveJSON(filepath.Join(h.dataDir, "workspaces", "data.json"), data)
}

// LoadWorkspaces loads workspaces from file
func (h *HTTPHandler) LoadWorkspaces() ([]Workspace, error) {
	var data WorkspaceData
	err := h.loadJSON(filepath.Join(h.dataDir, "workspaces", "data.json"), &data)
	if err != nil {
		return []Workspace{}, nil
	}
	return data.Workspaces, nil
}

// SaveCollections saves collections to file
func (h *HTTPHandler) SaveCollections(collections []Collection) error {
	data := CollectionData{Collections: collections}
	return h.saveJSON(filepath.Join(h.dataDir, "collections", "data.json"), data)
}

// LoadCollections loads collections from file
func (h *HTTPHandler) LoadCollections() ([]Collection, error) {
	var data CollectionData
	err := h.loadJSON(filepath.Join(h.dataDir, "collections", "data.json"), &data)
	if err != nil {
		return []Collection{}, nil
	}
	return data.Collections, nil
}

// SaveEnvironments saves environments to file
func (h *HTTPHandler) SaveEnvironments(environments []Environment) error {
	data := EnvironmentData{Environments: environments}
	return h.saveJSON(filepath.Join(h.dataDir, "environments", "data.json"), data)
}

// LoadEnvironments loads environments from file
func (h *HTTPHandler) LoadEnvironments() ([]Environment, error) {
	var data EnvironmentData
	err := h.loadJSON(filepath.Join(h.dataDir, "environments", "data.json"), &data)
	if err != nil {
		return []Environment{}, nil
	}
	return data.Environments, nil
}

// SaveHistory saves request history to file
func (h *HTTPHandler) SaveHistory(items []HistoryItem) error {
	data := HistoryData{Items: items}
	return h.saveJSON(filepath.Join(h.dataDir, "history", "data.json"), data)
}

// LoadHistory loads request history from file
func (h *HTTPHandler) LoadHistory() ([]HistoryItem, error) {
	var data HistoryData
	err := h.loadJSON(filepath.Join(h.dataDir, "history", "data.json"), &data)
	if err != nil {
		return []HistoryItem{}, nil
	}
	return data.Items, nil
}

// SaveSettings saves application settings to file
func (h *HTTPHandler) SaveSettings(settings Settings) error {
	return h.saveJSON(filepath.Join(h.dataDir, "settings", "data.json"), settings)
}

// LoadSettings loads application settings from file
func (h *HTTPHandler) LoadSettings() (*Settings, error) {
	var settings Settings
	err := h.loadJSON(filepath.Join(h.dataDir, "settings", "data.json"), &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// saveJSON saves data to a JSON file
func (h *HTTPHandler) saveJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

// loadJSON loads data from a JSON file
func (h *HTTPHandler) loadJSON(path string, data interface{}) error {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, data)
}

// GetDataDirectory returns the data directory path
func (h *HTTPHandler) GetDataDirectory() string {
	return h.dataDir
}