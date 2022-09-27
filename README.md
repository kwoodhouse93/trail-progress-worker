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
- `PROCESS_INTERVAL` - how often to check for activities that still need processing

## What does it do?

It subscribes to postgres notifications on the specified channel.

When something sends `NOTIFY {channel}`, the app will 'process' any unprocessed activities in the database. This involves calculating their intersections with long distance routes, and storing the results for use by the fronted. These are generally quite long-running operations. Processing only a handful of relevant activity/route overlaps can take on the order of minutes to calculate.

As a fallback for missed notifications, the app will also check for unprocessed activities every `PROCESS_INTERVAL`.
