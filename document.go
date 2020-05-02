package main

import (
	"strconv"
	"strings"
	"time"
)

type Document struct {
	DocumentID      string
	ReferenceID     string
	AssetIdentifier string
	InvoiceNumber   string
	TransactionType string
	EntityNumber    int
	EntityName      string
	Payee           string
	Payor           string
	InvoiceCurrency string
	InvoiceAmount   string
	PaymentTerms    int
	InvoiceDate     time.Time
	DueDate         time.Time
	RiskScore       string
	SchemaName		string
	CollateralValue string
	NFTToken        string
}

func toDocument(row []string) Document {
	entityNo, err := strconv.Atoi(row[4])
	checkErr(err)
	pt, err := strconv.Atoi(row[10])
	checkErr(err)
	id, err := time.Parse("1/2/2006", row[11])
	checkErr(err)
	dd, err := time.Parse("1/2/2006", row[12])
	checkErr(err)
	row[9] = strings.ReplaceAll(row[9], ",", "")
	row[14] = strings.ReplaceAll(row[14], ",", "")
	return Document{
		ReferenceID:     row[0],
		AssetIdentifier: row[1],
		InvoiceNumber:   row[2],
		TransactionType: row[3],
		EntityNumber:    entityNo,
		EntityName:      row[5],
		Payee:           row[6],
		Payor:           row[7],
		InvoiceCurrency: row[8],
		InvoiceAmount:   row[9],
		PaymentTerms:    pt,
		InvoiceDate:     id,
		DueDate:         dd,
		RiskScore:       row[13],
		CollateralValue: row[14],
		SchemaName: row[15],
	}
}
