package strava

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	activitiesPath = "/activities"
)

type GetActivityRequest struct {
	AccessToken       string
	ID                int
	IncludeAllEfforts bool
}

type GetActivityResponse struct {
	DetailedActivity
}

func (s *API) GetActivity(req GetActivityRequest) (*GetActivityResponse, error) {
	url := baseURL + activitiesPath + "/" + strconv.Itoa(req.ID)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+req.AccessToken)
	resp, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strava: failed to get activity: %s", string(respBody))
	}
	var response GetActivityResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
