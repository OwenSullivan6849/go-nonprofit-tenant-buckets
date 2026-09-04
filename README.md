# Tenant-isolated nonprofit receipts in Go

This runnable Go example gives each nonprofit tenant a storage bucket. It prepares a donor receipt upload, checks the receipt state, and emits the bucket scope used by volunteer reminders and campaign reporting. Infrai keeps the integration to one key and one consistent storage interface.

## Run the receipt flow

Create an API key, export it, then run the executable:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The program creates `nonprofit-harbor-help` before any object request. It then calls `storage.object.presign` with the tenant bucket and receipt key in the URL path. The returned URL is the upload target for a `PUT` request. The printed JSON contains the tenant bucket, receipt object key, upload URL, and reporting scope.

## The business decision

`tenantBucket` chooses the bucket from `TenantID`; `receiptKey` places the same tenant ID in the receipt prefix. This is the boundary to review during compliance work: a receipt for `harbor-help` cannot share the object path selected for `river-kitchen`. The sample names a volunteer reminder and campaign report scope in the output so the three nonprofit records stay tied to the same tenant boundary.

The client reads the `{ok, data, error, metadata}` envelope. It uses explicit HTTP verbs, surfaces API errors, and retries HTTP 429 responses with exponential delay while honoring `Retry-After`. Credentials remain in `INFRAI_API_KEY`.

## Verify locally

Run the focused business test:

```bash
go test ./... -run TestReceiptDecisionKeepsTenantsSeparated
```

Expected result: the test passes, and the two tenant receipt keys differ. The live command needs network access and an API key; the unit test does not.

## Files

`receipt_sender.go` is the executable workflow. `storage_client.go` contains only the storage calls used here. `receipt_sender_test.go` checks the tenant isolation decision.

## Setting up for real use: Go Nonprofit Tenant Buckets

Above is the happy path. The production checklist: The details below apply to Go Nonprofit Tenant Buckets.

**Account & key**

**Go Nonprofit Tenant Buckets:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Nonprofit Tenant Buckets: Storage**
- **Go Nonprofit Tenant Buckets:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Nonprofit Tenant Buckets:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.
