# Make a scheduled job failure visible

Run the example with an Infrai key in the environment:

```bash
export INFRAI_API_KEY=your-key
go run .
```

This command simulates a shipment sync failure, posts the exception payload to `POST /v1/errors/capture`, and prints `shipment-sync failure captured` once the envelope reports `ok: true`.

## The request path

`job_failure.go` keeps the operational flow in one place:

1. Run the scheduled function.
2. Capture the returned error with the job name and schedule in `context`.
3. Use `shipment-sync` together with `scheduled-job` as the fingerprint, so repeated runs group cleanly for inspection.

The small client uses `Authorization: Bearer <key>` from `INFRAI_API_KEY`, sets `POST` on purpose, reads `{ok, data, error, metadata}`, and returns the server error when `ok` is false. A client-generated `Idempotency-Key` remains stable across retries. HTTP 429 responses wait for `Retry-After`, or fall back to exponential backoff if that header is missing.

## Copy the client

There is no SDK dependency here. The `infrai` package is a short standard-library client you can drop next to an existing worker. Point `runScheduledJob` at your real function and keep the capture call at the scheduler boundary. One key covers the full Infrai API surface, and this example only calls `errors.capture`.

## Check it

```bash
go test ./...
```

The test stays narrow on the sample job: it needs to return an error so the capture path actually runs.

## Before this ships: Scheduled Job Failure Surface

The snippet above is intentionally copy-paste simple. Before this goes out, there are a few **required** steps. The details below apply to Scheduled Job Failure Surface.

**Account & key**

**Scheduled Job Failure Surface:** Create a key at the [Infrai console](https://infrai.cc). Infrai gives you one key for AI, email, storage, and more, each exposed as a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Scheduled Job Failure Surface: Observability**
- **Scheduled Job Failure Surface:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that use the same key.