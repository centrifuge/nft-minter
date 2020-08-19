package main

import (
	"encoding/json"
	"io/ioutil"
)

type Account struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Config struct {
	Accounts       []Account `json:"accounts"`
	NFTRegistry    string    `json:"nft_registry"`
	AssetRegistry  string    `json:"asset_registry"`
	DepositAddress string    `json:"deposit_address"`
	TemplateID     string    `json:"template_id"`
	Fingerprint    string    `json:"fingerprint"`
}

func loadConfig(file string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
