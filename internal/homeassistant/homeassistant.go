package homeassistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HomeAssistant struct {
	baseURL string
	token   string
	client  *http.Client
}

type Device struct {
	EntityID string `json:"entity_id" db:"entity_id"`
	Name     string `json:"name" db:"name"`
	Type     string `json:"type" db:"type"`
}

type ServiceRequest struct {
	EntityID string `json:"entity_id"`
}

type StateResponse struct {
	EntityID   string                 `json:"entity_id"`
	State      string                 `json:"state"`
	Attributes map[string]interface{} `json:"attributes"`
}

func New(baseURL, token string) *HomeAssistant {
	return &HomeAssistant{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ha *HomeAssistant) makeRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ha.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+ha.token)
	req.Header.Set("Content-Type", "application/json")

	return ha.client.Do(req)
}

func (ha *HomeAssistant) TurnOn(entityID string) error {
	deviceType := getDeviceType(entityID)
	endpoint := fmt.Sprintf("/api/services/%s/turn_on", deviceType)

	resp, err := ha.makeRequest("POST", endpoint, ServiceRequest{EntityID: entityID})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to turn on device: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (ha *HomeAssistant) TurnOff(entityID string) error {
	deviceType := getDeviceType(entityID)
	endpoint := fmt.Sprintf("/api/services/%s/turn_off", deviceType)

	resp, err := ha.makeRequest("POST", endpoint, ServiceRequest{EntityID: entityID})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to turn off device: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (ha *HomeAssistant) Toggle(entityID string) error {
	deviceType := getDeviceType(entityID)
	endpoint := fmt.Sprintf("/api/services/%s/toggle", deviceType)

	resp, err := ha.makeRequest("POST", endpoint, ServiceRequest{EntityID: entityID})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to toggle device: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (ha *HomeAssistant) GetState(entityID string) (*StateResponse, error) {
	endpoint := fmt.Sprintf("/api/states/%s", entityID)

	resp, err := ha.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get device state: %s - %s", resp.Status, string(body))
	}

	var state StateResponse
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (ha *HomeAssistant) PressButton(entityID string) error {
	endpoint := "/api/services/button/press"

	resp, err := ha.makeRequest("POST", endpoint, ServiceRequest{EntityID: entityID})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to press button: %s - %s", resp.Status, string(body))
	}

	return nil
}

// getDeviceType extracts the device type from entity_id (e.g., "light.living_room" -> "light")
func getDeviceType(entityID string) string {
	for i, c := range entityID {
		if c == '.' {
			return entityID[:i]
		}
	}
	return entityID
}
