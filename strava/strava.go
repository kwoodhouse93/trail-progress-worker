package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	baseURL          = "https://www.strava.com/api/v3"
	subscriptionPath = "/push_subscriptions"
)

type API struct {
	client       *http.Client
	clientID     int
	clientSecret string
}

func NewAPI(clientID int, clientSecret string) *API {
	return &API{
		client:       &http.Client{},
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

type CreateSubscriptionRequest struct {
	CallbackURL string
	VerifyToken string
}

type CreateSubscriptionResponse struct {
	ID int `json:"id"`
}

type createSubscriptionRequest struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CallbackURL  string `json:"callback_url"`
	VerifyToken  string `json:"verify_token"`
}

func (s *API) CreateSubscription(req CreateSubscriptionRequest) (*CreateSubscriptionResponse, error) {
	log.Println("strava: creating strava webhooks subscription")
	r := createSubscriptionRequest{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		CallbackURL:  req.CallbackURL,
		VerifyToken:  req.VerifyToken,
	}
	reqBody, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	url := baseURL + subscriptionPath
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava: failed to create subscription: %s", string(respBody))
	}
	var response CreateSubscriptionResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// TODO: Verify response type when there is a subscription
type ViewSubscriptionResponse struct {
	ID            int    `json:"id"`
	ResourceState int    `json:"resource_state"`
	ApplicationID int    `json:"application_id"`
	CallbackURL   string `json:"callback_url"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func (s *API) ViewSubscription() ([]ViewSubscriptionResponse, error) {
	q := url.Values{
		"client_id":     {strconv.Itoa(s.clientID)},
		"client_secret": {s.clientSecret},
	}
	url, err := url.Parse(baseURL + subscriptionPath)
	if err != nil {
		return nil, err
	}
	url.RawQuery = q.Encode()
	resp, err := s.client.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava: failed to get subscription: %s", string(respBody))
	}
	var response []ViewSubscriptionResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type DeleteSubscriptionRequest struct {
	ID int
}

type deleteSubscriptionRequest struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (s *API) DeleteSubscription(req DeleteSubscriptionRequest) error {
	// DELETE https://www.strava.com/api/v3/push_subscriptions/id
	url := baseURL + subscriptionPath + "/" + strconv.Itoa(req.ID)
	r := deleteSubscriptionRequest{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
	}
	reqBody, err := json.Marshal(r)
	if err != nil {
		return err
	}
	deleteRequest, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	deleteRequest.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(deleteRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("strava: failed to delete subscription: %s", string(respBody))
	}
	return nil
}
