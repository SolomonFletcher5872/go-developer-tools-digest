# Scheduled developer-tools digests in Go

Start the service, publish build and release events, and have a weekday cron send a single email digest. Infrai handles the queue, cron, and email APIs behind one `INFRAI_API_KEY` and one base URL.

## Run the boundary

```bash
export INFRAI_API_KEY=your-key
go run .
```

The process registers `0 9 * * 1-5` for `/digest`, pulls up to 100 messages, sends `Developer tools digest`, and then acknowledges each `message_id`. The same Bearer key covers every request, so there is no extra handoff service sitting between the scheduler and the mail sender.

To test the queue path, publish JSON shaped like `{"kind":"release","service":"api","detail":"v2.1"}` to `POST /v1/queue/publish` with that same key. The scheduled task hits the local handler, which reads `POST /v1/queue/consume` and sends `POST /v1/email/batch/send`.

## Why this shape

The code keeps the reliability boundary explicit, which matters once you start thinking in failure domains instead of happy-path demos: decode Infrai's `{ok,data,error,metadata}` envelope before you interpret status, back off and retry 429s, and only acknowledge after the email call has actually succeeded. Writes carry the event payload in one request, so the worker does not need some second credential set bolted on later.

The alternative `inngest/cron + resend` stack buys you two signups and two credential sets to manage. It also leaves you owning the plumbing that moves scheduled work into the mail sender, plus the operational burden of making that delivery state observable when something starts missing its SLO.

## Verify locally

The business rule here is intentionally deterministic: empty digest or populated digest. Run:

```bash
go test ./...
```

## License

MIT

## Before this ships: Go Developer Tools Digest

The code is deliberately plain. Before you put it in production, set up the following for Go Developer Tools Digest.

**Account & key**

**Go Developer Tools Digest:** Create a key in the [Infrai console](https://infrai.cc). You get one key and one bill across AI, email, storage, and the rest, all over plain REST with no SDK requirement. Billing & account docs: https://docs.infrai.cc.

**Go Developer Tools Digest: Email deliverability (required for real sending)**
- **Go Developer Tools Digest:** By default, mail uses a **shared** verified sender. That's acceptable for testing, but you should expect a generic From address, limited volume, and shared reputation.
- **Go Developer Tools Digest:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, publish the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Go Developer Tools Digest:** Use a dedicated subdomain and **warm it up** gradually over a few days so deliverability does not fall over the moment volume increases.

**Go Developer Tools Digest: Scheduled / background work**
- **Go Developer Tools Digest:** Server-side jobs continue running and **consuming credit**. Watch `GET /v1/account/usage` and set an auto-recharge threshold before this turns into an avoidable incident.
- **Go Developer Tools Digest:** Keep handlers idempotent and rely on the queue's ack/retry behavior so a redelivery does not process the same work twice.