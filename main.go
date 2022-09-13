package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-gitlab-framework"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"

	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

type options struct {
	service  liboptions.ServiceOptions
	gitlab   liboptions.GitLabOptions
	endpoint string
}

func (o *options) Validate() error {
	if err := o.service.Validate(); err != nil {
		return err
	}

	if err := o.gitlab.Validate(); err != nil {
		return err
	}

	if o.endpoint == "" {
		return errors.New("missing gitlab-endpoint")
	}

	return nil
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.gitlab.AddFlags(fs)
	o.service.AddFlags(fs)

	fs.StringVar(&o.endpoint, "gitlab-endpoint", "", "the endpoint of gitlab.")

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

	// load gitlab token
	gitlabToken, err := ioutil.ReadFile(o.gitlab.TokenPath)
	if err != nil {
		log.Errorf("read gitlab token failed, err:%s", err.Error())

		return
	}

	// gitlab client
	cli, err := newGitlabClient(gitlabToken, o.endpoint)
	if err != nil {
		log.Errorf("new gitlab client failed, err:%s", err.Error())

		return
	}

	// obs client
	oc := &cfg.OBSConfig
	obsClient, err := obs.New(oc.AccessKey, oc.SecretKey, oc.Endpoint)
	if err != nil {
		log.Errorf("new obs client failed, err:%s", err.Error())

		return
	}

	// sync service
	service := sync.NewSyncService(
		&cfg.SyncConfig, obsClient, nil,
		func(pid string) (string, error) {
			return getLastestCommit(cli, pid)
		},
	)

	r := newRobot(
		"root", string(gitlabToken),
		cfg.AccessHmac, cfg.AccessEndpoint,
		service,
	)

	framework.Run(r, o.service.Port, o.service.GracePeriod)
}

func newGitlabClient(token []byte, host string) (*gitlab.Client, error) {
	tc := string(token)
	opts := gitlab.WithBaseURL(host)

	return gitlab.NewOAuthClient(tc, opts)
}

func getLastestCommit(cli *gitlab.Client, pid string) (string, error) {
	v, _, err := cli.Commits.ListCommits(pid, nil, func(req *retryablehttp.Request) error {
		v := req.URL.Query()
		v.Add("per_page", "1")
		v.Add("page=1", "1")
		req.URL.RawQuery = v.Encode()

		return nil
	})

	if err != nil || len(v) == 0 {
		return "", err
	}

	return v[0].ID, nil
}
