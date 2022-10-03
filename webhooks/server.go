package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/kwoodhouse93/trail-progress-worker/store"
	"github.com/kwoodhouse93/trail-progress-worker/strava"
	"github.com/pkg/errors"
)

type Server struct {
	server      *http.Server
	stravaAPI   *strava.API
	store       *store.Store
	verifyToken string
}

func NewServer(addr string, stravaAPI *strava.API, store *store.Store, verifyToken string) *Server {
	s := &http.Server{
		Addr: addr,
	}

	return &Server{
		server:      s,
		stravaAPI:   stravaAPI,
		store:       store,
		verifyToken: verifyToken,
	}
}

func (s Server) Serve() error {
	log.Println("webhooks: starting webhook server")
	s.server.Handler = s.handleWebhooks()
	return s.server.ListenAndServe()
}

func (s Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s Server) handleWebhooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleSubscriptionValidation()(w, r)
			return
		}
		if r.Method == http.MethodPost {
			s.handleEvent()(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

type subscriptionValidationResponse struct {
	Challenge string `json:"hub.challenge"`
}

func (s Server) handleSubscriptionValidation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GET https://mycallbackurl.com?hub.verify_token=STRAVA&hub.challenge=15f7d1a91c1f40f8a748fd134752feb3&hub.mode=subscribe
		q := r.URL.Query()
		if q.Get("hub.mode") != "subscribe" {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("webhooks: received subscription validation request with invalid hub.mode")
			return
		}
		if q.Get("hub.verify_token") != s.verifyToken {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("webhooks: received subscription validation request with incorrect hub.verify_token:", q.Get("hub.verify_token"))
			return
		}
		if !q.Has("hub.challenge") {
			w.WriteHeader(http.StatusBadRequest)
			log.Println("webhooks: received subscription validation request with no hub.challenge")
			return
		}

		resp := subscriptionValidationResponse{
			Challenge: q.Get("hub.challenge"),
		}
		respBody, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("webhooks: failed to marshal subscription validation response: %v", err)
			return
		}
		w.Write(respBody)
		w.Header().Set("Content-Type", "application/json")
	}
}

type aspectType string

const (
	AspectTypeCreate aspectType = "create"
	AspectTypeUpdate aspectType = "update"
	AspectTypeDelete aspectType = "delete"
)

type objectType string

const (
	ObjectTypeActivity objectType = "activity"
	ObjectTypeAthlete  objectType = "athlete"
)

type updateField string

const (
	UpdateFieldTitle      updateField = "title"
	UpdateFieldType       updateField = "type"
	UpdateFieldPrivate    updateField = "private"
	UpdateFieldAuthorized updateField = "authorized"
)

type updates map[updateField]interface{}

// Returns (title, nil)  if `"title"` was present in `updates`
// Returns (nil, nil) if `"title"` was not present in `updates`
// Returns (nil, err) if `"title"` was present but was not a string as expected
func (u updates) Title() (*string, error) {
	value, ok := u[UpdateFieldTitle]
	if !ok {
		return nil, nil
	}
	title, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("webhooks: failed to parse title update: %v", value)
	}
	return &title, nil
}

// Returns (type, nil)  if `"type"` was present in `updates`
// Returns (nil, nil) if `"type"` was not present in `updates`
// Returns (nil, err) if `"type"` was present but was not a string as expected
func (u updates) Type() (*string, error) {
	value, ok := u[UpdateFieldType]
	if !ok {
		return nil, nil
	}
	typeVal, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("webhooks: failed to parse type update: %v", value)
	}
	return &typeVal, nil
}

// Returns (private, nil)  if `"private"` was present in `updates`
// Returns (nil, nil) if `"private"` was not present in `updates`
// Returns (nil, err) if `"private"` was present but was not "true" or "false" as expected
func (u updates) Private() (*bool, error) {
	value, ok := u[UpdateFieldPrivate]
	if !ok {
		return nil, nil
	}
	private, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("webhooks: failed to parse private update: %v", value)
	}
	return &private, nil
}

// Returns (true, nil)  if `"authorized": "false"` was received
// Returns (false, nil) if `"authorized"` was not present in `updates`
// Returns (false, err) if `"authorized"` was present but was not `"false"` as expected
func (u updates) Authorized() (bool, error) {
	value, ok := u[UpdateFieldAuthorized]
	if !ok {
		return false, nil
	}
	authorized, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("webhooks: failed to parse authorized update: %v", value)
	}
	if authorized != "false" {
		return false, fmt.Errorf("webhooks: received unexpected authorized update: %v", authorized)
	}
	return true, nil
}

type webhookRequest struct {
	ObjectType     objectType `json:"object_type"`
	ObjectID       int        `json:"object_id"`
	AspectType     aspectType `json:"aspect_type"`
	Updates        updates    `json:"updates"`
	OwnerID        int        `json:"owner_id,omitempty"`
	SubscriptionID int        `json:"subscription_id"`
	EventTime      int        `json:"event_time"`
}

func (s Server) handleEvent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("webhooks: failed to read request body: %v", err)
			return
		}
		var req webhookRequest
		err = json.Unmarshal(reqBody, &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("webhooks: failed to unmarshal request body: %v", err)
			return
		}
		log.Printf("webhooks: received event: %+v\n", req)

		switch req.ObjectType {
		case ObjectTypeActivity:
			switch req.AspectType {
			case AspectTypeCreate:
				err := s.handleCreateActivity(r.Context(), req)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("webhooks: failed to handle create activity event: %v", err)
					return
				}
			case AspectTypeUpdate:
				err := s.handleUpdateActivity(r.Context(), req)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("webhooks: failed to handle update activity event: %v", err)
					return
				}
			case AspectTypeDelete:
				err := s.handleDeleteActivity(r.Context(), req)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("webhooks: failed to handle delete activity event: %v", err)
					return
				}
			default:
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("webhooks: received activity event with unexpected aspect type: %v", req.AspectType)
				return
			}
		case ObjectTypeAthlete:
			switch req.AspectType {
			case AspectTypeUpdate:
				err := s.handleUpdateAthlete(r.Context(), req)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("webhooks: failed to handle update athlete event: %v", err)
					return
				}
			default:
				w.WriteHeader(http.StatusBadRequest)
				log.Printf("webhooks: received athlete event with unexpected aspect type: %v", req.AspectType)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("webhooks: received event with unexpected object type: %v", req.ObjectType)
			return
		}
	}
}

func (s Server) handleCreateActivity(ctx context.Context, req webhookRequest) error {
	token, err := s.AccessToken(ctx, req.OwnerID)
	if err != nil {
		return errors.Wrap(err, "webhooks: failed to get access token")
	}

	resp, err := s.stravaAPI.GetActivity(strava.GetActivityRequest{
		AccessToken: token.Token,
		ID:          req.ObjectID,
	})
	if err != nil {
		return errors.Wrap(err, "webhooks: failed to get activity")
	}
	log.Println("webhooks: got new activity, ID:", resp.ID)

	err = s.store.StoreActivity(ctx, req.OwnerID, resp.DetailedActivity)
	if err != nil {
		return errors.Wrap(err, "webhooks: failed to store activity")
	}
	return nil
}

func (s Server) handleUpdateActivity(ctx context.Context, req webhookRequest) error {
	return errors.New("unimplemented")
}

func (s Server) handleDeleteActivity(ctx context.Context, req webhookRequest) error {
	return errors.New("unimplemented")
}

func (s Server) handleUpdateAthlete(ctx context.Context, req webhookRequest) error {
	return errors.New("unimplemented")
}

// Gets the user's access token, refreshing it if necessary.
func (s Server) AccessToken(ctx context.Context, athleteID int) (*store.AccessToken, error) {
	token, err := s.store.GetAccessToken(ctx, athleteID)
	if err != nil {
		return nil, errors.Wrap(err, "webhooks: failed to get access token")
	}
	log.Println("webhooks: got access token:", token)

	if token.IsExpired() {
		log.Println("webhooks: token expired, refreshing")
		refreshToken, err := s.store.GetRefreshToken(ctx, athleteID)
		if err != nil {
			return nil, errors.Wrap(err, "webhooks: failed to get refresh token")
		}
		resp, err := s.stravaAPI.RefreshToken(strava.RefreshTokenRequest{
			RefreshToken: refreshToken.Token,
		})
		if err != nil {
			return nil, errors.Wrap(err, "webhooks: failed to refresh token")
		}
		log.Println("webhooks: refreshed token:", resp)

		err = s.store.StoreTokens(
			ctx,
			athleteID,
			resp.AccessToken,
			resp.RefreshToken,
			time.UnixMilli(int64(resp.ExpiresAt*1000)),
		)
		if err != nil {
			return nil, errors.Wrap(err, "webhooks: failed to store tokens")
		}
		token = &store.AccessToken{
			Token:     resp.AccessToken,
			ExpiresAt: time.UnixMilli(int64(resp.ExpiresAt * 1000)),
		}
		log.Println("webhooks: stored refreshed token:", token.Token)
	}

	return token, nil
}
