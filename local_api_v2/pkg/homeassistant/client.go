package homeassistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents Home Assistant API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Status     bool
}

// NewClient creates a new Home Assistant client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Status: false,
	}
}

func (c *Client) IsConnected() bool {
	resp, err := c.HTTPClient.Get(c.BaseURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return true
	}
	return false
}

// State represents a Home Assistant entity state
type State struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
}

// InputNumberEntity represents an input_number entity in Home Assistant
type InputNumberEntity struct {
	EntityID     string  `json:"entity_id"`
	FriendlyName string  `json:"friendly_name"`
	Value        float64 `json:"value"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Step         float64 `json:"step"`
	Unit         string  `json:"unit_of_measurement,omitempty"`
}

// GetStates retrieves all states from Home Assistant
func (c *Client) GetStates() ([]State, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/states", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var states []State
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return states, nil
}

// GetInputNumbers retrieves all input_number entities
func (c *Client) GetInputNumbers() ([]InputNumberEntity, error) {
	states, err := c.GetStates()
	if err != nil {
		return nil, err
	}

	var inputNumbers []InputNumberEntity
	for _, state := range states {
		if strings.HasPrefix(state.EntityID, "input_number.") {
			value := parseFloat(state.State)

			entity := InputNumberEntity{
				EntityID:     state.EntityID,
				FriendlyName: getStringAttribute(state.Attributes, "friendly_name", state.EntityID),
				Value:        value,
				Min:          getFloatAttribute(state.Attributes, "min", 0),
				Max:          getFloatAttribute(state.Attributes, "max", 100),
				Step:         getFloatAttribute(state.Attributes, "step", 1),
				Unit:         getStringAttribute(state.Attributes, "unit_of_measurement", ""),
			}

			inputNumbers = append(inputNumbers, entity)
		}
	}

	return inputNumbers, nil
}

// SetInputNumber sets the value of an input_number entity
func (c *Client) SetInputNumber(entityID string, value float64) error {
	payload := map[string]interface{}{
		"entity_id": entityID,
		"value":     value,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/services/input_number/set_value", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetState gets the current state of a specific entity
func (c *Client) GetState(entityID string) (*State, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/states/"+entityID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("entity %s not found", entityID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var state State
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &state, nil
}

// Helper functions
func parseFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0
	}
}

func getFloatAttribute(attrs map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := attrs[key]; ok {
		return parseFloat(val)
	}
	return defaultValue
}

func getStringAttribute(attrs map[string]interface{}, key string, defaultValue string) string {
	if val, ok := attrs[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}
