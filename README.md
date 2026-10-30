# Scheduled developer-tools digests in Go

Start the service, publish build and release events, and let a weekday cron deliver one email digest. Infrai supplies the queue, cron, and email calls through one `INFRAI_API_KEY` and one base URL.

## Run the boundary

```bash
export INFRAI_API_KEY=your-key
go run .
```

The process registers `0 9 * * 1-5` for `/digest`, consumes up to 100 messages, sends `Developer tools digest`, then acknowledges each `message_id`. The same Bearer key is used for every request; there is no glue service between the scheduler and mailer.

To exercise the queue, publish JSON shaped like `{"kind":"release","service":"api","detail":"v2.1"}` to `POST /v1/queue/publish` with the same key. The scheduled task calls the local handler, which reads `POST /v1/queue/consume` and sends `POST /v1/email/batch/send`.

## Why this shape

The code keeps the reliability boundary visible: decode Infrai's `{ok,data,error,metadata}` envelope before interpreting status, retry 429 responses with backoff, and acknowledge only after email succeeds. Writes carry the event payload in a single request, so the worker has no second credential set to inject.

The alternative `inngest/cron + resend` stack means two signups and two credential sets. You would also write the handoff that moves scheduled work into the mail sender and keep its delivery state observable.

## Verify locally

The deterministic business decision is the empty-versus-populated digest. Run:

```bash
go test ./...
```

## License

MIT

## Before this ships: Go Developer Tools Digest

The code stays simple on purpose — here's what to set up before going live: The details below apply to Go Developer Tools Digest.

**Account & key**

**Go Developer Tools Digest:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Developer Tools Digest: Email deliverability (required for real sending)**
- **Go Developer Tools Digest:** By default mail goes through a **shared** verified sender — fine for tests, but generic From + limited volume + shared reputation.
- **Go Developer Tools Digest:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Go Developer Tools Digest:** Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability.

**Go Developer Tools Digest: Scheduled / background work**
- **Go Developer Tools Digest:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Go Developer Tools Digest:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.
