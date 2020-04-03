package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	AutoRecover    bool   `json:"autoRecover"`
	ServerDir      string `json:"serverDir"`
	Jarfile        string `json:"jarfile"`
	Xmx            string `json:"xmx"`
	Xms            string `json:"xms"`
	BackupDir      string `json:"backupDir"`
	WorldName      string `json:"worldName"`
	BackupInterval int    `json:"backupInterval"`
	PruneAge       int    `json:"pruneAge"`
}

func LoadConfig(path string) (Config, error) {
	cfg, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(cfg, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}
