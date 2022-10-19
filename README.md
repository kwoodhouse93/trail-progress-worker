# trail-progress-worker
Background worker for running geographical processing on Strava activities for [kwoodhouse93/trail-progress](https://github.com/kwoodhouse93/trail-progress).

## Run

```sh
go run ./...
```

## Build

```sh
docker build . -t trail-progress-worker:latest
```

## Deploy

Put the docker container somewhere...

Don't forget to specify any env vars:

- `POSTGRES_CONNECTION_STRING` - should be kept secret
- `POSTGRES_LISTEN_CHANNEL` - the channel to listen for notifications on
- `PROCESSOR_INTERVAL` - how often to check for activities that still need processing
- `PROCESSOR_CONCURRENCY` - how many workers to run concurrently
- `PROCESSOR_BATCH_SIZE` - how many activity/route pairs to process in each transaction
- `STRAVA_CLIENT_ID` - see https://developers.strava.com/ for more info
- `STRAVA_CLIENT_SECRET` - should also be kept secret
- `STRAVA_CALLBACK_URL` - the URL this server can be reached on

This repo comes with config for deploying on [fly.io](https://fly.io/) - see [`/fly.toml`](https://github.com/kwoodhouse93/trail-progress-worker/blob/main/fly.toml).

## Continuous delivery

[`.github/workflows`](https://github.com/kwoodhouse93/trail-progress-worker/tree/main/.github/workflows) contains a workflow for deploying on fly with each push to `main`.

## What does it do?

2 main functions:
1. Background processing of Strava activities
2. Listening for Strava webhooks

### Background processing
It subscribes to postgres notifications on the specified channel.

When something sends `NOTIFY {channel}`, the app will 'process' any unprocessed activities in the database. This involves calculating their intersections with long distance routes, and storing the results for use by the fronted.

These can be quite long-running operations. Processing only a handful of relevant activity/route overlaps can take on the order of minutes to calculate. For this reason, we process them in batches and can run multiple workers concurrently. For more info, see [`/store/README.md`](https://github.com/kwoodhouse93/trail-progress-worker/blob/main/store/README.md)

As a fallback for missed notifications, each worker goroutine will also check for unprocessed activities every `PROCESS_INTERVAL`.

They will also check for work to do on startup - handy if you want to manually trigger processing by running a copy locally, or by restarting the deployed service.

### Strava webhooks

Strava implements a webhook system. Apps can subscribe to receive real-time updates for certain events, eliminating the need for polling.

For more info, see the Strava developer docs on webhooks:  
https://developers.strava.com/docs/webhooks/

This service will attempt to establish a webhook subscription on startup, and will try to unsubscribe itself on shutdown.

Webhooks can be received for new activities, changes to activities, deletion of activities, and deauthorisation by an athlete.
