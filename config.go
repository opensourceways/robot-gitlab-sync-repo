package main

import (
	"github.com/opensourceways/community-robot-lib/utils"

	"github.com/opensourceways/robot-gitlab-sync-repo/infrastructure/obsimpl"
	"github.com/opensourceways/robot-gitlab-sync-repo/infrastructure/platformimpl"
	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type configuration struct {
	// AccessEndpoint is used to send back the message.
	AccessEndpoint string `json:"access_endpoint" required:"true"`
	AccessHmac     string `json:"access_hmac"     required:"true"`

	Sync   sync.Config         `json:"sync"`
	OBS    obsimpl.Config      `json:"obs"`
	Gitlab platformimpl.Config `json:"gitlab"`
}

func (cfg *configuration) configItems() []interface{} {
	return []interface{}{
		&cfg.Sync,
		&cfg.OBS,
		&cfg.Gitlab,
	}
}

func (cfg *configuration) Validate() error {
	if _, err := utils.BuildRequestBody(cfg, ""); err != nil {
		return err
	}

	items := cfg.configItems()

	for _, i := range items {
		if v, ok := i.(configValidate); ok {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cfg *configuration) SetDefault() {
	items := cfg.configItems()

	for _, i := range items {
		if v, ok := i.(configSetDefault); ok {
			v.SetDefault()
		}
	}
}

func loadConfig(file string) (cfg configuration, err error) {
	if err = utils.LoadFromYaml(file, &cfg); err != nil {
		return
	}

	cfg.SetDefault()

	err = cfg.Validate()

	return
}
