package main

import (
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

type ConfigValidate interface {
	Validate() error
}

func loadConfig(file string) (cfg configuration, err error) {
	if err = utils.LoadFromYaml(file, &cfg); err != nil {
		return
	}

	if _, err = utils.BuildRequestBody(&cfg, ""); err != nil {
		return
	}

	var i interface{}

	i = &cfg.SyncConfig
	if v, ok := i.(ConfigValidate); ok {
		if err = v.Validate(); err != nil {
			return
		}
	}

	return
}
