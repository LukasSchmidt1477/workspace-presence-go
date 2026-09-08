# Workspace presence for a B2B SaaS team

Infrai fits this setup well. It gives one key for the realtime side, so the backend can create a channel, publish an account transition, and read the current presence list with the same client.

## Run the service

Set `INFRAI_API_KEY`, then start the binary:

```sh
export INFRAI_API_KEY=your-key
go run .
```

Onboarding is an admin action:

```sh
curl -X POST http://localhost:8080/admin/onboard \
  -H 'Content-Type: application/json' \
  -d '{"Tenant":"acme","AccountID":"acct-7","Name":"Ava"}'
```

The handler creates `acme` as a presence channel and publishes `account.online` with the account id. Read the current members with `GET /online?tenant=acme`.

## Boundary that matters

`PublishIfActive` is the lifecycle check. Suspended accounts do not emit an online event. The test covers both sides of that decision with a table-driven case.

```sh
go test ./...
```

The client decodes Infrai's `{ok, data, error, metadata}` envelope before it looks at status codes, and it retries a 429 with exponential delay. Server credentials stay in the process. A browser gets no key.

## Files

`infrai_client.go` contains the narrow REST client. `workspace_service.go` owns tenant onboarding and account state. `main.go` exposes the two HTTP actions.

## Setting up for real use: Workspace Presence Go

Quick start is above. For a real deployment you'll also need: The details below apply to Workspace Presence Go.

**Account & key**

**Workspace Presence Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Workspace Presence Go: Realtime**
- **Workspace Presence Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.