package main

import "github.com/opensourceways/robot-gitlab-sync-repo/sync"

type configuration struct {
	SyncConfig sync.Config `json:"sync"`
	OBSConfig  OBSConfig   `json:"obs"`
}

func (c *configuration) Validate() error {
	return nil
}

func (c *configuration) SetDefault() {

}

type OBSConfig struct {
	AccessKey string `json:"access_key" required:"true"`
	SecretKey string `json:"secret_key" required:"true"`
	Endpoint  string `json:"endpoint" required:"true"`
}
