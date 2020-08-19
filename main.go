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

	// out := make(chan bool)
	// go initScanRead(out)

	// alice creates template
	// <-out
	docID := config.TemplateID
	fingerprint := config.Fingerprint
	if docID == "" {
		fmt.Println("Alice creating a template....")
		docID, err = createDocument(alice.ID, alice.URL, nil)
		checkErr(err)
		fmt.Println("TemplateID:", docID)

		// alice adds roles and rules
		// <-out
		fmt.Println("Alice creating compute field rules with bob as collaborator...")
		roleID, err := createRole(alice.ID, bob.ID, docID, alice.URL)
		checkErr(err)
		fmt.Println("RoleID containing Bob:", roleID)
		// <-out
		err = createComputeRule(alice.ID, alice.URL, docID, roleID, "./simple_average.wasm")
		checkErr(err)
		fmt.Println("Alice created compute field rule")

		// alice commits the template
		// <-out
		fmt.Println("Alice committing template...")
		fingerprint, err = commitDocument(alice.ID, alice.URL, docID)
		checkErr(err)
	}

	fmt.Println("Template ID:", docID)
	fmt.Println("Template fingerprint:", fingerprint)

	// alice clones template and creates a draft document
	fmt.Println("Alice creating a document from template")
	docID, err = cloneDocument(alice.ID, alice.URL, docID, nil)
	checkErr(err)
	fingerprint, err = commitDocument(alice.ID, alice.URL, docID)
	checkErr(err)
	docID, err = updateDocument(alice.ID, alice.URL, docID, initAttributes(alice.ID, docID))
	checkErr(err)
	fmt.Println("Alice committing document...")
	fingerprint, err = commitDocument(alice.ID, alice.URL, docID)
	checkErr(err)
	fmt.Println("Anchored Document:", docID)
	fmt.Println("Document fingerprint:", fingerprint)

	// fetch attribute
	// <-out
	fmt.Println("Fetching Compute field result attribute...")
	attr, err := fetchAttribute(alice.ID, docID, alice.URL, "result")
	checkErr(err)
	risk, value := riskAndValue(attr)
	fmt.Printf("Resultant attribute value: risk = %s value = %s\n", risk, value)

	// bob updates the document
	// <-out
	fmt.Println("Bob updating the document with attributes required for compute fields...")
	docID, err = updateDocument(bob.ID, bob.URL, docID, computeAttributes())
	checkErr(err)
	fmt.Println("Updated document:", docID)

	// bob commits the document
	// <-out
	fmt.Println("Bob committing document...")
	fingerprint, err = commitDocument(bob.ID, bob.URL, docID)
	checkErr(err)
	fmt.Println("Anchored document:", docID)
	fmt.Println("Document fingerprint:", fingerprint)

	// fetch attribute
	// <-out
	fmt.Println("Fetching Compute field result attribute...")
	attr, err = fetchAttribute(alice.ID, docID, alice.URL, "result")
	checkErr(err)
	risk, value = riskAndValue(attr)
	fmt.Printf("Resultant attribute value: risk = %s value = %s\n", risk, value)

	// mint NFT
	fmt.Println("Alice mints NFT...")
	token, err := mintNFT(docID, alice.ID, alice.URL, config.NFTRegistry, config.AssetRegistry, config.DepositAddress)
	checkErr(err)
	fmt.Println("NFT tokenID: ", token)
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
