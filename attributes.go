package main

import (
	"time"
)

type AttributeRequest struct {
	Type  string `json:"type" enums:"integer,decimal,string,bytes,timestamp,monetary"`
	Value string `json:"value"`
}

func initAttributes(id, docID string) map[string]AttributeRequest {
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
			Value: time.Now().Format(time.RFC3339Nano),
		},
		"AssetValue": {
			Type:  "decimal",
			Value: "1100",
		},
		"RiskScore": {
			Type:  "integer",
			Value: "1",
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
		"value1": {
			Type:  "integer",
			Value: "1000",
		},
		"value2": {
			Type:  "integer",
			Value: "950",
		},
		"value3": {
			Type:  "integer",
			Value: "1100",
		},
	}
}
