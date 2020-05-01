package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func fetchSigningKey(url, id string) (string, error) {
	url = fmt.Sprintf("%s/v1/accounts/%s", url, id)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var response struct {
		SigningKeyPair struct {
			Pub string `json:"pub"`
		} `json:"signing_key_pair"`
	}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return "", err
	}

	return response.SigningKeyPair.Pub, nil
}

type AttributeRequest struct {
	Type  string `json:"type" enums:"integer,decimal,string,bytes,timestamp,monetary"`
	Value string `json:"value"`
}

func toAttributes(doc Document, did string) map[string]AttributeRequest {
	attrs := map[string]AttributeRequest{
		"reference_id": {
			Type:  "string",
			Value: doc.ReferenceID,
		},
		"_schema": {
			Type: "string",
			Value: doc.SchemaName,
		},
		"invoice_nr": {
			Type:  "string",
			Value: doc.InvoiceNumber,
		},
		"AssetIdentifier": {
			Type:  "bytes",
			Value: doc.AssetIdentifier,
		},
		"transaction_type": {
			Type:  "string",
			Value: doc.TransactionType,
		},
		"entity_nr": {
			Type:  "integer",
			Value: strconv.Itoa(doc.EntityNumber),
		},
		"entity_name": {
			Type:  "string",
			Value: doc.EntityName,
		},
		"payee": {
			Type:  "string",
			Value: doc.Payee,
		},
		"payor": {
			Type:  "string",
			Value: doc.Payor,
		},
		"invoice_amount": {
			Type:  "decimal",
			Value: doc.InvoiceAmount,
		},
		"currency": {
			Type:  "string",
			Value: doc.InvoiceCurrency,
		},
		"payment_terms": {
			Type:  "integer",
			Value: strconv.Itoa(doc.PaymentTerms),
		},
		"invoice_date": {
			Type:  "timestamp",
			Value: doc.InvoiceDate.Format(time.RFC3339Nano),
		},
		"MaturityDate": {
			Type:  "timestamp",
			Value: doc.DueDate.Format(time.RFC3339Nano),
		},
		"risk_score": {
			Type:  "string",
			Value: doc.RiskScore,
		},
		"AssetValue": {
			Type:  "decimal",
			Value: doc.CollateralValue,
		},
		"Originator": {
			Type:  "bytes",
			Value: did,
		},
	}

	return attrs
}

func createDocument(doc Document, config Config) (Document, error) {
	url := fmt.Sprintf("%s/v2/documents", config.NodeURL)
	doc, err := updateDocument(doc, url, "POST", config, false, http.StatusCreated, false)
	if err != nil {
		return Document{}, err
	}

	url = fmt.Sprintf("%s/v2/documents/%s", config.NodeURL, doc.DocumentID)
	doc, err = updateDocument(doc, url, "PATCH", config, true, http.StatusOK, false)
	if err != nil {
		return Document{}, err
	}

	url = fmt.Sprintf("%s/v2/documents/%s/commit", config.NodeURL, doc.DocumentID)
	doc, err = updateDocument(doc, url, "POST", config, false, http.StatusAccepted, true)
	return doc, err
}

func updateDocument(doc Document, url string, method string, config Config, patch bool, status int, commit bool) (Document, error) {
	payload := map[string]interface{}{
		"scheme": "generic",
		"data":   map[string]interface{}{},
	}

	if patch {
		payload["attributes"] = toAttributes(doc, config.CentrifugeID)
	}

	d, err := json.Marshal(payload)
	if err != nil {
		return Document{}, err
	}

	var r io.Reader
	r = bytes.NewReader(d)
	if commit {
		r = nil
	}

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return Document{}, err
	}

	req.Header.Add("authorization", config.CentrifugeID)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return Document{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != status {
		return Document{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response struct {
		Header struct {
			JobID      string `json:"job_id"`
			DocumentID string `json:"document_id"`
		}
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Document{}, err
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return Document{}, err
	}

	doc.DocumentID = response.Header.DocumentID
	if commit {
		err = waitForTransactionSuccess(config.NodeURL, config.CentrifugeID, response.Header.JobID)
	}
	return doc, err
}

func waitForTransactionSuccess(url, did, jobID string) error {
	url = fmt.Sprintf("%s/v1/jobs/%s", url, jobID)
	c := new(http.Client)
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		req.Header.Add("authorization", did)
		req.Header.Add("accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		res, err := c.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}

		var response struct {
			Message string `json:"message"`
			Status  string `json:"status"`
		}

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &response)
		if err != nil {
			return err
		}

		if response.Status == "pending" {
			time.Sleep(1 * time.Second)
			continue
		}

		if response.Status == "success" {
			return nil
		}

		return errors.New(response.Message)
	}
}

func mintNFT(doc Document, config Config) (Document, error) {
	pfs := []string{
		"cd_tree.attributes[0xe24e7917d4fcaf79095539ac23af9f6d5c80ea8b0d95c9cd860152bff8fdab17].byte_val",
		"cd_tree.attributes[0xcd35852d8705a28d4f83ba46f02ebdf46daf03638b40da74b9371d715976e6dd].byte_val",
		"cd_tree.attributes[0xbbaa573c53fa357a3b53624eb6deab5f4c758f299cffc2b0b6162400e3ec13ee].byte_val",
		"cd_tree.attributes[0xe5588a8a267ed4c32962568afe216d4ba70ae60576a611e3ca557b84f1724e29].byte_val",
	}

	sk, err := fetchSigningKey(config.NodeURL, config.CentrifugeID)
	if err != nil {
		return Document{}, err
	}
	skb, err := hexutil.Decode(sk)
	if err != nil {
		return Document{}, err
	}

	idb, err := hexutil.Decode(config.CentrifugeID)
	if err != nil {
		return Document{}, err
	}

	pub := GetAddress(skb)
	key := append(idb, pub[:]...)
	pfs = append(pfs, fmt.Sprintf("%s.signatures[%s]", "signatures_tree", hexutil.Encode(key)))

	url := fmt.Sprintf("%s/v1/nfts/registries/%s/mint", config.NodeURL, config.NFTRegistry)
	payload := map[string]interface{}{
		"asset_manager_address": config.AssetContract,
		"deposit_address":       config.NFTDepositAddress,
		"document_id":           doc.DocumentID,
		"proof_fields":          pfs,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return Document{}, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return Document{}, err
	}

	req.Header.Add("authorization", config.CentrifugeID)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, err := new(http.Client).Do(req)
	if err != nil {
		return Document{}, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return Document{}, fmt.Errorf("unexpected error: %d", res.StatusCode)
	}

	var response struct {
		Header struct {
			JobID string `json:"job_id"`
		} `json:"header"`
		TokenID string `json:"token_id"`
	}
	d := json.NewDecoder(res.Body)
	err = d.Decode(&response)
	if err != nil {
		return Document{}, err
	}

	err = waitForTransactionSuccess(config.NodeURL, config.CentrifugeID, response.Header.JobID)
	if err != nil {
		return Document{}, err
	}

	doc.NFTToken = response.TokenID
	return doc, nil
}

func GetAddress(publicKey []byte) [32]byte {
	hash := crypto.Keccak256(publicKey[1:])
	address := hash[12:]
	return AddressTo32Bytes(address)
}
func AddressTo32Bytes(address []byte) [32]byte {
	address32Byte := [32]byte{}
	for i := 1; i <= common.AddressLength; i++ {
		address32Byte[32-i] = address[common.AddressLength-i]
	}
	return address32Byte
}
