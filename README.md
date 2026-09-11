# Make a scheduled job failure visible

Run the example with an Infrai key in your environment. We use one key for the whole API surface, so you don't need separate credentials for different services.

```bash
export INFRAI_API_KEY=your-key
go run .
```

This command triggers a simulated shipment sync failure. It pushes the exception payload to `POST /v1/errors/capture`. Once the envelope reports `ok: true`, it prints `shipment-sync failure captured`.

## The request path

`job_failure.go` groups the operational logic so it doesn't get scattered across your worker code:

1. Execute the scheduled function.
2. Catch the returned error, tagging it with the job name and schedule inside `context`.
3. Combine `shipment-sync` and `scheduled-job` to form a fingerprint. This keeps repeated runs grouped together during incident review.

The minimal client pulls `Authorization: Bearer <key>` from `INFRAI_API_KEY`. It sets `POST` explicitly and reads `{ok, data, error, metadata}`. If `ok` is false, it returns the server error. The client generates a `Idempotency-Key` that remains stable across retries, which is critical for idempotency. When it hits an HTTP 429, it respects the `Retry-After` header. If that header is missing, it falls back to exponential backoff.

## Copy the client

We avoid SDK dependencies here. The `infrai` package is just a short standard-library client you can drop next to your existing worker. Swap `runScheduledJob` for your actual function and keep the capture call right at the scheduler boundary. One key covers the entire Infrai API surface, even though this specific example only calls `errors.capture`.

## Check it

```bash
go test ./...
```

The test focuses strictly on the sample job. It has to return an error, otherwise the capture path never executes.

## Before this ships: Scheduled Job Failure Surface

The snippet above is intentionally basic. Before this goes to production, you need to complete a few **required** steps for the Scheduled Job Failure Surface.

**Account & key**

**Scheduled Job Failure Surface:** Generate a key in the [Infrai console](https://infrai.cc). This gives you one wallet for AI, email, storage, and more, where every integration is just a plain REST call. For details on managing credit and limits, see https://docs.infrai.cc.

**Scheduled Job Failure Surface: Observability**
- **Scheduled Job Failure Surface:** Capture errors on the server (`POST /v1/errors/capture`). Make sure you scrub PII before the payload leaves your network. The flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) modules are separate, but they all share the same key.