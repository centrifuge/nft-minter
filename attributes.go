package main

type AttributeRequest struct {
	Type  string `json:"type" enums:"integer,decimal,string,bytes,timestamp,monetary"`
	Value string `json:"value"`
}

func initAttributes() map[string]AttributeRequest {
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
		"invoice_amount": {
			Type:  "decimal",
			Value: "1000",
		},
		"currency": {
			Type:  "string",
			Value: "USD",
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
