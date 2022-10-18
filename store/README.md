# Process versions

The way we're using PostGIS uses a lot of compute resources. It has to load big GPS tracks and do some CPU-heavy computation to calculate buffers, intersections, etc.

This document explains some of the rationale behind the various process versions found in this directory.

## V1

V1 fetches all unprocessed activity/route pairs and processes them in one large batch. This can take several minutes to run depending on how many activities there are and how much they overlap with routes.

The downside here is that, due to foreign key constraints and the necessary use of a 'repeatable read' isolation level, the athlete will be locked while processing their activities. This causes issues elsewhere, especially with features like account deletion, since it will take far too long to acquire the lock needed to delete the athlete record.

For my account:

- Processing time: ~1.2 minutes
- Locking: blocks a lot of other DB write activity for the entire processing time

## V2

V2 attempts to address the shortcomings of V1 by only processing a single activity/route pair at a time. This should reduce the impact of locking tables/rows for long periods.

The obvious downside is we now have a new transaction, and possibly a new connection (we do use pgx pools to help manage connections) for each activity/route pair. My account has over 500 activities, so even with just 2 routes, that's already over 1,000 pairs. This introduces _a lot_ of overhead for managing transactions compared with V1.

Thus, processing time is severely impacted.

For my account:

- Processing time: >3 minutes
- Locking: Rarely more than a few seconds

### Concurrency

There's one more advantage of V2.

It should enable concurrent processing of many pairs at once. This does work, and it speeds processing time back up to ~1 minute.

However, the main issue I've run into when deploying this against a Supabase database is the memory limits. Free tier accounts can use up to 1GB of RAM before postgres will be unceremoniously killed.

Running more than two processors at a time will almost always trigger OOM. Even using just two processors will sometimes use too much memory. Using multiple connections increases memory usage substantially, from ~500MB max to well over 1GB.

## V3

V3 is an attempt to find a happy medium by processing activity/route pairs in batches. For example, using a batch size of 10 reduces the overhead caused by running 1000s of transactions by an order of magnitude.

It also seems to reduce memory usage - possibly due to the lower number of transactions. I can't verify the details, but it seems plausible.

For my account:
- Processing time: ~ 0.9 minutes
- Locking: Rarely more than ~5 seconds

This feels like a good compromise for now.

We could still unlock a whole lot more performance by using a Postgres server with higher memory limits. This would let us keep more connections open and process more activity/route pairs concurrently.

However, higher performance servers do not come cheap, so as long as this remains a non-commercial hobby project, the aim is to maximise performance within the constraints of a freely available server.

In fact, that Supabase even offer a server with this level of performance for free is already pretty sweet!
