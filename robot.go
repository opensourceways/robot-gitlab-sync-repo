package main

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	sdk "github.com/xanzy/go-gitlab"

	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

const botName = "sync_repo"

type iClient interface {
}

func newRobot(cli iClient, gc func() (*configuration, error)) *robot {
	return &robot{cli: cli, getConfig: gc}
}

type robot struct {
	getConfig func() (*configuration, error)
	cli       iClient
	root      string
	service   sync.SyncService
}

func (bot *robot) HandlePushEvent(e *sdk.PushEvent, log *logrus.Entry) error {
	repoName := e.Project.Name
	repoType := ""
	if strings.HasPrefix(repoName, "project") {
		repoType = "project"
	} else if strings.HasPrefix(repoName, "model") {
		repoType = "model"
	} else if strings.HasPrefix(repoName, "dataset") {
		repoType = "dataset"
	} else {
		return errors.New("unknown repo type")
	}

	url := strings.Replace(e.Project.GitHTTPURL, "://", bot.root, 1)

	if e.Before == "" {
		// no need to handle the first commit
		return nil
	}

	v := sync.RepoInfo{
		Owner:    e.Project.Namespace,
		RepoId:   e.ProjectID,
		RepoURL:  url,
		RepoName: repoName,
		RepoType: repoType,
	}

	if err := bot.service.Sync(&v); err == nil {
		return nil
	}

	// send back the request

	return nil
}
