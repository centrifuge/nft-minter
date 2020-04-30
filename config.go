package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"
)

type Config struct {
	CentrifugeID      string `json:"centrifuge_id"`
	NodeURL           string `json:"node_url"`
	AssetContract     string `json:"asset_contract"`
	NFTRegistry       string `json:"nft_registry"`
	NFTDepositAddress string `json:"nft_deposit_address"`
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

	if config.CentrifugeID == "" {
		return Config{}, errors.New("centrifuge ID is not set in config")
	}

	if config.NodeURL == "" {
		return Config{}, errors.New("nodeURl is not set in config")
	} else {
		config.NodeURL = strings.TrimSuffix(config.NodeURL, "/")
	}

	if config.AssetContract == "" {
		return Config{}, errors.New("asset contract is not set in config")
	}

	if config.NFTRegistry == "" {
		return Config{}, errors.New("nft registry is not set in config")
	}

	return config, nil
}
