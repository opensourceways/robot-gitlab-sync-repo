package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
	sdk "github.com/xanzy/go-gitlab"

	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

const botName = "sync_repo"

func newRobot(user, token, hmac, endpoint string, s sync.SyncService) *robot {
	return &robot{
		root:     fmt.Sprintf("://%s:%s@", user, token),
		hmac:     hmac,
		endpoint: endpoint,
		service:  s,
		hc:       utils.HttpClient{MaxRetries: 3},
	}
}

type robot struct {
	root     string
	hmac     string
	endpoint string
	hc       utils.HttpClient
	service  sync.SyncService
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

	return bot.sendBack(e)
}

func (bot *robot) sendBack(e *sdk.PushEvent) error {
	body, err := utils.JsonMarshal(e)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost, bot.endpoint, bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	h := &req.Header
	h.Add("Content-Type", "application/json")
	h.Add("User-Agent", botName)
	h.Add("X-Gitlab-Event", "System Hook")
	h.Add("X-Gitlab-Token", bot.hmac)
	h.Add("X-Gitlab-Event-UUID", "73ed8438-1119-4bb8-ae9d-0180c88ef168")

	return bot.hc.ForwardTo(req, nil)
}
