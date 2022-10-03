package strava

import (
	"net/http"
)

const (
	baseURL = "https://www.strava.com/api/v3"
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
