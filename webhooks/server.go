package webhooks

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/kwoodhouse93/trail-progress-worker/strava"
)

type Server struct {
	server      *http.Server
	stravaAPI   *strava.API
	verifyToken string
}

func NewServer(addr string, stravaAPI *strava.API, verifyToken string) *Server {
	s := &http.Server{
		Addr: addr,
	}

	return &Server{
		server:      s,
		stravaAPI:   stravaAPI,
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
