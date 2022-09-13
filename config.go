package main

import (
	"encoding/json"

	"github.com/opensourceways/community-robot-lib/utils"

	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

type configuration struct {
	AccessEndpoint string `json:"access_endpoint" required:"true"`
	AccessHmac     string `json:"access_hmac" required:"true"`

	SyncConfig sync.Config `json:"sync"`
	OBSConfig  OBSConfig   `json:"obs"`
}

type OBSConfig struct {
	AccessKey string `json:"access_key" required:"true"`
	SecretKey string `json:"secret_key" required:"true"`
	Endpoint  string `json:"endpoint" required:"true"`
}

func loadConfig(file string) (cfg configuration, err error) {
	if err = utils.LoadFromYaml(file, &cfg); err != nil {
		return
	}

	// fix
	_, err = json.Marshal(&cfg)

	return
}
