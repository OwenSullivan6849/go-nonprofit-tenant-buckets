package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type donorReceipt struct {
	TenantID    string `json:"tenant_id"`
	ReceiptID   string `json:"receipt_id"`
	DonorName   string `json:"donor_name"`
	AmountCents int    `json:"amount_cents"`
}

func main() {
	client, err := NewStorageClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	receipt := donorReceipt{TenantID: "harbor-help", ReceiptID: "rcpt-2026-001", DonorName: "A. Chen", AmountCents: 12500}
	bucket := tenantBucket(receipt.TenantID)
	key := receiptKey(receipt.TenantID, receipt.ReceiptID)
	signed, err := client.PresignPut(bucket, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	status, err := client.Head(bucket, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	output, _ := json.MarshalIndent(map[string]any{
		"bucket": bucket, "receipt_key": key, "upload_url": signed.URL,
		"receipt_ready": status.Found, "volunteer_reminder": "send after receipt upload",
		"campaign_report_scope": bucket,
	}, "", "  ")
	fmt.Println(string(output))
}
