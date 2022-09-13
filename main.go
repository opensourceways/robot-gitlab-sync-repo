package main

import (
	"errors"
	"flag"
	"os"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-gitlab-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type options struct {
	service  liboptions.ServiceOptions
	token    liboptions.GitLabOptions
	endpoint string
}

func (o *options) Validate() error {
	if err := o.service.Validate(); err != nil {
		return err
	}

	return o.token.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.token.AddFlags(fs)
	o.service.AddFlags(fs)
	fs.StringVar(&o.endpoint, "gitlab-endpoint", "", "the endpoint of gitlab.")

	fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.token.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	agent := config.NewConfigAgent(func() config.Config {
		return &configuration{}
	})

	if err := agent.Start(o.service.ConfigFile); err != nil {
		logrus.WithError(err).Errorf("start config:%s", o.service.ConfigFile)
		return
	}

	defer agent.Stop()

	c := gitlabclient.NewGitlabClient(
		secretAgent.GetTokenGenerator(o.token.TokenPath),
		o.endpoint,
	)

	r := newRobot(c, func() (*configuration, error) {
		_, cfg := agent.GetConfig()
		if c, ok := cfg.(*configuration); ok {
			return c, nil
		}

		return nil, errors.New("can't convert to configuration")
	})

	framework.Run(r, o.service.Port, o.service.GracePeriod)
}

func newGitlabClient(getToken func() []byte, host string) (*gitlab.Client, error) {
	tc := string(getToken())
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
