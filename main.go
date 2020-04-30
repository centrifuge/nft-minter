package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/opentracing/opentracing-go/log"
)

type Response struct {
	doc Document
	err error
}

func main() {
	args := os.Args
	if len(args) < 3 {
		panic("need 2 argument")
	}

	cfile := args[1]
	config, err := loadConfig(cfile)
	checkErr(err)
	file := args[2]
	f, err := os.Open(file)
	checkErr(err)
	cr := csv.NewReader(f)
	rows, err := cr.ReadAll()
	checkErr(err)
	rows = rows[1:]
	var documents []Document
	resp := make(chan Response)
	for _, r := range rows {
		doc := toDocument(r)
		go func(document Document) {
			document, err := createDocument(document, config)
			resp <- Response{
				doc: document,
				err: err,
			}
		}(doc)
	}

	for i := 0; i < len(rows); i++ {
		r := <-resp
		if r.err != nil {
			log.Error(err)
			continue
		}
		r.doc, err = mintNFT(r.doc, config)
		if err != nil {
			log.Error(err)
			continue
		}
		documents = append(documents, r.doc)
		fmt.Printf("InvoiceNumber: %s DocumentIdentifier: %s NFTToken: %s\n", r.doc.InvoiceNumber, r.doc.DocumentID, r.doc.NFTToken)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
