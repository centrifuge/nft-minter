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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func createDocument(id, nodeURL string, attrs map[string]AttributeRequest) (string, error) {
	url := fmt.Sprintf("%s/v2/documents", nodeURL)
	return documentAction(id, url, "", "POST", http.StatusCreated, attrs)
}

func updateDocument(id, nodeURL, docID string, attrs map[string]AttributeRequest) (string, error) {
	url := fmt.Sprintf("%s/v2/documents", nodeURL)
	return documentAction(id, url, docID, "POST", http.StatusCreated, attrs)
}

func cloneDocument(id, nodeURL, docID string, attrs map[string]AttributeRequest) (string, error) {
	url := fmt.Sprintf("%s/v2/documents/%s/clone", nodeURL, docID)
	return documentAction(id, url, docID, "POST", http.StatusCreated, attrs)
}

func documentAction(id, url, docID, method string, status int, attrs map[string]AttributeRequest) (string, error) {
	payload := map[string]interface{}{
		"scheme":       "generic",
		"data":         map[string]interface{}{},
		"document_id":  docID,
		"attributes":   attrs,
		"write_access": []string{id},
	}

	d, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	var response struct {
		Header struct {
			JobID      string `json:"job_id"`
			DocumentID string `json:"document_id"`
		}
	}

	err = makeCall(id, url, method, status, bytes.NewReader(d), &response)
	return response.Header.DocumentID, err
}

func commitDocument(id, nodeURL, docID string) (string, error) {
	var response struct {
		Header struct {
			JobID       string `json:"job_id"`
			DocumentID  string `json:"document_id"`
			Fingerprint string `json:"fingerprint"`
		}
	}
	url := fmt.Sprintf("%s/v2/documents/%s/commit", nodeURL, docID)
	err := makeCall(id, url, "POST", http.StatusAccepted, nil, &response)
	if err != nil {
		return "", err
	}
	return response.Header.Fingerprint, waitForTransactionSuccess(nodeURL, id, response.Header.JobID)
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
				"attribute_labels":       []string{"RiskScore", "AssetValue"},
				"target_attribute_label": "result",
			},
		},
		"attribute_rules": {
			{
				"key_label": "AssetValue",
				"role_id":   roleID,
			},
			{
				"key_label": "RiskScore",
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
	url := fmt.Sprintf("%s/v2/documents/%s/committed", nodeURL, docID)
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
	url = fmt.Sprintf("%s/v2/jobs/%s", url, jobID)
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
			Finished bool `json:"finished"`
			Tasks    []struct {
				RunnerFunc string      `json:"runnerFuncs"` // name of the runnerFuncs
				Result     interface{} `json:"result"`      // result after the task run
				Error      string      `json:"error"`       // error after task run
			}
		}

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &response)
		if err != nil {
			return err
		}

		if !response.Finished {
			time.Sleep(1 * time.Second)
			continue
		}

		if message := response.Tasks[len(response.Tasks)-1].Error; message != "" {
			return errors.New(message)
		}

		return nil
	}
}

func riskAndValue(result string) (risk, value *big.Int) {
	d := hexutil.MustDecode(result)
	risk = new(big.Int).SetBytes(d[:16])
	value = new(big.Int).SetBytes(d[16:])
	return risk, value
}

func fetchSigningKey(url, id string) (string, error) {
	url = fmt.Sprintf("%s/v2/accounts/%s", url, id)
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

func mintNFT(docID, id, nodeURL, nftRegistry, assetContract, depositAddress string) (tokenID string, err error) {
	pfs := []string{
		"cd_tree.attributes[0xe24e7917d4fcaf79095539ac23af9f6d5c80ea8b0d95c9cd860152bff8fdab17].byte_val",
		"cd_tree.attributes[0xcd35852d8705a28d4f83ba46f02ebdf46daf03638b40da74b9371d715976e6dd].byte_val",
		"cd_tree.attributes[0xbbaa573c53fa357a3b53624eb6deab5f4c758f299cffc2b0b6162400e3ec13ee].byte_val",
		"cd_tree.attributes[0xe5588a8a267ed4c32962568afe216d4ba70ae60576a611e3ca557b84f1724e29].byte_val",
	}

	sk, err := fetchSigningKey(nodeURL, id)
	if err != nil {
		return tokenID, err
	}
	skb, err := hexutil.Decode(sk)
	if err != nil {
		return tokenID, err
	}

	idb, err := hexutil.Decode(id)
	if err != nil {
		return tokenID, err
	}

	pub := GetAddress(skb)
	key := append(idb, pub[:]...)
	pfs = append(pfs, fmt.Sprintf("%s.signatures[%s]", "signatures_tree", hexutil.Encode(key)))

	url := fmt.Sprintf("%s/v2/nfts/registries/%s/mint", nodeURL, nftRegistry)
	payload := map[string]interface{}{
		"asset_manager_address": assetContract,
		"deposit_address":       depositAddress,
		"document_id":           docID,
		"proof_fields":          pfs,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return tokenID, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return tokenID, err
	}

	req.Header.Add("authorization", id)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	res, err := new(http.Client).Do(req)
	if err != nil {
		return tokenID, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		return tokenID, fmt.Errorf("unexpected error: %d", res.StatusCode)
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
		return tokenID, err
	}

	err = waitForTransactionSuccess(nodeURL, id, response.Header.JobID)
	if err != nil {
		return tokenID, err
	}

	return response.TokenID, nil
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
