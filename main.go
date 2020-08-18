package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	config, err := loadConfig("./config.json")
	checkErr(err)
	alice := config.Accounts[0]
	bob := config.Accounts[1]

	out := make(chan bool)
	go initScanRead(out)

	// alice creates document
	<-out
	fmt.Println("Alice creating a draft document...")
	docID, err := createDocument(alice.ID, alice.URL, "", initAttributes())
	checkErr(err)
	fmt.Println("DocumentID:", docID)

	// alice adds roles and rules
	<-out
	fmt.Println("Alice creating compute field rules with bob as collaborator...")
	roleID, err := createRole(alice.ID, bob.ID, docID, alice.URL)
	checkErr(err)
	fmt.Println("RoleID containing Bob:", roleID)
	<-out
	err = createComputeRule(alice.ID, alice.URL, docID, roleID, "./simple_average.wasm")
	checkErr(err)
	fmt.Println("Alice created compute field rule")

	// alice commits the document
	<-out
	fmt.Println("Alice committing document...")
	err = commitDocument(alice.ID, alice.URL, docID)
	checkErr(err)
	fmt.Println("Anchored document:", docID)

	// fetch attribute
	<-out
	fmt.Println("Fetching Compute field result attribute...")
	attr, err := fetchAttribute(alice.ID, docID, alice.URL, "result")
	checkErr(err)
	risk, value := riskAndValue(attr)
	fmt.Printf("Resultant attribute value: risk = %s value = %s\n", risk, value)

	// bob updates the document
	<-out
	fmt.Println("Bob updating the document with attributes required for compute fields...")
	docID, err = createDocument(bob.ID, bob.URL, docID, computeAttributes())
	checkErr(err)
	fmt.Println("Updated document:", docID)

	// bob commits the document
	<-out
	fmt.Println("Bob committing document...")
	err = commitDocument(bob.ID, bob.URL, docID)
	checkErr(err)
	fmt.Println("Anchored document:", docID)

	// fetch attribute
	<-out
	fmt.Println("Fetching Compute field result attribute...")
	attr, err = fetchAttribute(alice.ID, docID, alice.URL, "result")
	checkErr(err)
	risk, value = riskAndValue(attr)
	fmt.Printf("Resultant attribute value: risk = %s value = %s\n", risk, value)
}

func initScanRead(out chan<- bool) {
	s := bufio.NewScanner(os.Stdin)
	for {
		out <- s.Scan()
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
