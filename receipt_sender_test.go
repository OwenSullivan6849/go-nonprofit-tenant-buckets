package main

import "testing"

func TestReceiptDecisionKeepsTenantsSeparated(t *testing.T) {
	first := receiptKey("harbor-help", "rcpt-001")
	second := receiptKey("river-kitchen", "rcpt-001")
	if first != "receipts/harbor-help/rcpt-001.json" {
		t.Fatalf("unexpected first key: %s", first)
	}
	if first == second {
		t.Fatal("receipts from separate tenants must use separate object keys")
	}
	if tenantBucket("Harbor-Help") != "nonprofit-harbor-help" {
		t.Fatal("bucket naming must be stable")
	}
}
