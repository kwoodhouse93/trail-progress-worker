package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const tokenPath = "/oauth/token"

type RefreshTokenRequest struct {
	RefreshToken string
}

type refreshTokenRequest struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	GrantType    string `json:"grant_type"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (s *API) RefreshToken(req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	r := refreshTokenRequest{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		RefreshToken: req.RefreshToken,
		GrantType:    "refresh_token",
	}
	reqBody, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	url := baseURL + tokenPath
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava: failed to refresh token: %s", string(respBody))
	}
	var response RefreshTokenResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
