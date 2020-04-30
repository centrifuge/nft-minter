package main

import (
	"strconv"
	"strings"
	"time"
)

type Document struct {
	DocumentID      string
	ReferenceID     string
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
	CollateralValue string
	NFTToken        string
}

func toDocument(row []string) Document {
	entityNo, err := strconv.Atoi(row[3])
	checkErr(err)
	pt, err := strconv.Atoi(row[9])
	checkErr(err)
	id, err := time.Parse("1/2/2006", row[10])
	checkErr(err)
	dd, err := time.Parse("1/2/2006", row[11])
	checkErr(err)
	row[8] = strings.ReplaceAll(row[8], ",", "")
	row[13] = strings.ReplaceAll(row[13], ",", "")
	return Document{
		ReferenceID:     row[0],
		InvoiceNumber:   row[1],
		TransactionType: row[2],
		EntityNumber:    entityNo,
		EntityName:      row[4],
		Payee:           row[5],
		Payor:           row[6],
		InvoiceCurrency: row[7],
		InvoiceAmount:   row[8],
		PaymentTerms:    pt,
		InvoiceDate:     id,
		DueDate:         dd,
		RiskScore:       row[12],
		CollateralValue: row[13],
	}
}
