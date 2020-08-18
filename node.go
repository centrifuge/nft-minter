package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func createDocument(id, nodeURL, docID string, attrs map[string]AttributeRequest) (string, error) {
	payload := map[string]interface{}{
		"scheme":      "generic",
		"data":        map[string]interface{}{},
		"document_id": docID,
		"attributes":  attrs,
	}

	d, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v2/documents", nodeURL)
	var response struct {
		Header struct {
			JobID      string `json:"job_id"`
			DocumentID string `json:"document_id"`
		}
	}

	err = makeCall(id, url, "POST", http.StatusCreated, bytes.NewReader(d), &response)
	return response.Header.DocumentID, err
}

func commitDocument(id, nodeURL, docID string) error {
	var response struct {
		Header struct {
			JobID      string `json:"job_id"`
			DocumentID string `json:"document_id"`
		}
	}
	url := fmt.Sprintf("%s/v2/documents/%s/commit", nodeURL, docID)
	err := makeCall(id, url, "POST", http.StatusAccepted, nil, &response)
	if err != nil {
		return err
	}
	return waitForTransactionSuccess(nodeURL, id, response.Header.JobID)
}

func createRole(alice, bob, docID, nodeURL string) (roleID string, err error) {
	url := fmt.Sprintf("%s/v2/documents/%s/roles", nodeURL, docID)
	d, err := json.Marshal(map[string]interface{}{
		"collaborators": []string{bob},
		"key":           "random_key",
	})
	if err != nil {
		return roleID, err
	}

	var response struct {
		ID string `json:"id"`
	}

	err = makeCall(alice, url, "POST", http.StatusOK, bytes.NewReader(d), &response)
	return response.ID, err
}

func createComputeRule(id, nodeURL, docID, roleID, file string) error {
	url := fmt.Sprintf("%s/v2/documents/%s/transition_rules", nodeURL, docID)
	wasm, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	d, err := json.Marshal(map[string][]map[string]interface{}{
		"compute_fields_rules": {
			{
				"wasm":                   hexutil.Encode(wasm),
				"attribute_labels":       []string{"value1", "value2", "value3"},
				"target_attribute_label": "result",
			},
		},
		"attribute_rules": {
			{
				"key_label": "value1",
				"role_id":   roleID,
			},
			{
				"key_label": "value2",
				"role_id":   roleID,
			},
			{
				"key_label": "value3",
				"role_id":   roleID,
			},
		},
	})
	if err != nil {
		return err
	}

	return makeCall(id, url, "POST", http.StatusOK, bytes.NewReader(d), &struct{}{})
}

func fetchAttribute(id, docID, nodeURL, attr string) (string, error) {
	url := fmt.Sprintf("%s/v1/documents/%s", nodeURL, docID)
	var response struct {
		Attributes map[string]struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	}

	err := makeCall(id, url, "GET", http.StatusOK, nil, &response)
	if err != nil {
		return "", err
	}

	for k, a := range response.Attributes {
		if k == attr {
			return a.Value, nil
		}
	}

	return "", errors.New("attribute not found")
}

func makeCall(id, url, method string, status int, reader io.Reader, response interface{}) error {
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}

	req.Header.Add("authorization", id)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != status {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &response)
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

func riskAndValue(result string) (risk, value *big.Int) {
	d := hexutil.MustDecode(result)
	risk = new(big.Int).SetBytes(d[:16])
	value = new(big.Int).SetBytes(d[16:])
	return risk, value
}
