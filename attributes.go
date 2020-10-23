package main

import (
	"time"
)

type AttributeRequest struct {
	Type  string `json:"type" enums:"integer,decimal,string,bytes,timestamp,monetary"`
	Value string `json:"value"`
}

func initAttributes(id, docID string) map[string]AttributeRequest {
	md := time.Now()
	return map[string]AttributeRequest{
		"reference_id": {
			Type:  "string",
			Value: "CF-001",
		},
		"invoice_nr": {
			Type:  "string",
			Value: "9500667307",
		},
		"entity_name": {
			Type:  "string",
			Value: "TechCargo",
		},
		"payee": {
			Type:  "string",
			Value: "Seeboard",
		},
		"payor": {
			Type:  "string",
			Value: "Daimler",
		},
		"currency": {
			Type:  "string",
			Value: "USD",
		},
		"MaturityDate": {
			Type:  "timestamp",
			Value: md.Format(time.RFC3339Nano),
		},
		"Originator": {
			Type:  "bytes",
			Value: id,
		},
		"AssetIdentifier": {
			Type:  "bytes",
			Value: docID,
		},
	}
}

func computeAttributes() map[string]AttributeRequest {
	return map[string]AttributeRequest{
		"AssetValue": {
			Type:  "integer",
			Value: "1100",
		},
		"RiskScore": {
			Type:  "integer",
			Value: "1",
		},
	}
}
