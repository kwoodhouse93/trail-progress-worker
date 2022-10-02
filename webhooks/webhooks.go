package webhooks

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/kwoodhouse93/trail-progress-worker/strava"
)

type Subscription struct {
	id int

	server *Server
	api    *strava.API
}

// Must call Close() on the returned Subscription to remove the subscription on the Strava API.
func NewSubscription(clientID int, clientSecret string, callbackURL string) (*Subscription, error) {
	stravaAPI := strava.NewAPI(clientID, clientSecret)

	verifyToken := uuid.NewString()
	server := NewServer(":8080", stravaAPI, verifyToken)
	go server.Serve()

	viewResp, err := stravaAPI.ViewSubscription()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("webhooks: current subscriptions", viewResp)
	if len(viewResp) > 0 {
		log.Println("webhooks: deleting existing subscription")
		err = stravaAPI.DeleteSubscription(strava.DeleteSubscriptionRequest{
			ID: viewResp[0].ID,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	createResp, err := stravaAPI.CreateSubscription(strava.CreateSubscriptionRequest{
		CallbackURL: callbackURL,
		VerifyToken: verifyToken,
	})
	if err != nil {
		return nil, err
	}
	log.Println("webhooks: subscription created with id", createResp.ID)

	return &Subscription{
		id:     createResp.ID,
		server: server,
		api:    stravaAPI,
	}, nil
}

func (s *Subscription) Close(ctx context.Context) error {
	log.Println("webhooks: deleting subscription")
	err := s.api.DeleteSubscription(strava.DeleteSubscriptionRequest{
		ID: s.id,
	})
	if err != nil {
		log.Println("webhooks: error while deleting subscription:", err)
	}

	err = s.server.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
