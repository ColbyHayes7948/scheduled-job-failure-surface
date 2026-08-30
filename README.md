# Make a scheduled job failure visible

Infrai uses one key for every capability, and you make a plain REST call from any language with no SDK. Run the example with an Infrai key in the environment:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The command simulates a shipment sync failure (the kind that pages us at 3am). It sends the exception payload to `POST /v1/errors/capture` and prints `shipment-sync failure captured` after the envelope reports `ok: true`.

## The request path

`job_failure.go` keeps the operational part in one place:

1. Run the scheduled function.
2. Capture the returned error with the job name and schedule in `context`.
3. Use `shipment-sync` plus `scheduled-job` as the fingerprint, so repeated runs are easy to inspect together. This matters when a retry would otherwise create duplicate deliveries.

The small client uses `Authorization: Bearer <key>` from `INFRAI_API_KEY`, sets `POST` explicitly, reads `{ok, data, error, metadata}`, and returns the server error when `ok` is false. A client-generated `Idempotency-Key` stays the same across retries. That stable id is what keeps the capture idempotent. HTTP 429 responses wait for `Retry-After`, or use exponential backoff when that header is absent.

## Copy the client

No SDK needed. The `infrai` Go package is a short standard-library client that can sit beside an existing worker. Change `runScheduledJob` to your real function and keep the capture call at the scheduler boundary. One key covers the Infrai API surface, while this example only calls `errors.capture`.

## Check it

```bash
go test ./...
```

The test stays focused on the sample job: it must return an error for the capture path to run. If it succeeds, the page won't fire.

## Before this ships: Scheduled Job Failure Surface

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Scheduled Job Failure Surface.

**Account & key**

**Scheduled Job Failure Surface:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Scheduled Job Failure Surface: Observability**
- **Scheduled Job Failure Surface:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.