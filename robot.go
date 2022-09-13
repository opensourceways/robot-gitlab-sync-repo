package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	sdk "github.com/xanzy/go-gitlab"

	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

const botName = "sync_repo"

func newRobot(token string, s sync.SyncService) *robot {
	return &robot{
		root:    fmt.Sprintf("://root:%s@", token),
		service: s,
	}
}

type robot struct {
	root    string
	service sync.SyncService
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

	v := sync.RepoInfo{
		Owner:    e.Project.Namespace,
		RepoId:   strconv.Itoa(e.ProjectID),
		RepoURL:  url,
		RepoName: repoName,
		RepoType: repoType,
	}

	if err := bot.service.SyncRepo(&v); err == nil {
		return nil
	}

	// send back the request

	return nil
}
