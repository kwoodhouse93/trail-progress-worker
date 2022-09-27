# trail-progress-worker
Background worker for running geographical processing on Strava activities for github.com/kwoodhouse93/trail-progress.

## Run

```sh
go run ./...
```

## Build

```sh
docker build .
```

## Deploy

Put the docker container somewhere...

Don't forget to specify any env vars:

- `POSTGRES_CONNECTION_STRING` - should be kept secret
- `POSTGRES_LISTEN_CHANNEL` - the channel to listen for notifications on
