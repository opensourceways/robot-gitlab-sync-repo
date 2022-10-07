package main

import (
	"flag"
	"os"

	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-gitlab-framework"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-gitlab-sync-repo/infrastructure/obsimpl"
	"github.com/opensourceways/robot-gitlab-sync-repo/infrastructure/platformimpl"
	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

type options struct {
	service liboptions.ServiceOptions
}

func (o *options) Validate() error {
	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(botName)
	log := logrus.NewEntry(logrus.StandardLogger())

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	// load config
	cfg, err := loadConfig(o.service.ConfigFile)
	if err != nil {
		log.Errorf("load config failed, err:%s", err.Error())

		return
	}

	gitlab, err := platformimpl.NewPlatform(&cfg.Gitlab)
	if err != nil {
		log.Errorf("init gitlab platform failed, err:%s", err.Error())

		return
	}

	// obs service
	obsService, err := obsimpl.NewOBS(&cfg.OBS)
	if err != nil {
		log.Errorf("init obs service failed, err:%s", err.Error())

		return
	}

	// sync service
	service := sync.NewSyncService(
		&cfg.Sync, obsService, nil, gitlab,
	)

	r := newRobot(
		cfg.AccessHmac, cfg.AccessEndpoint, service,
	)

	framework.Run(r, o.service.Port, o.service.GracePeriod)
}
