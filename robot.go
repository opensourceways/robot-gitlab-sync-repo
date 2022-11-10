package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
	sdk "github.com/xanzy/go-gitlab"

	"github.com/opensourceways/robot-gitlab-sync-repo/domain"
	"github.com/opensourceways/robot-gitlab-sync-repo/sync"
)

const botName = "sync_repo"

func newRobot(hmac, endpoint string, s sync.SyncService) *robot {
	return &robot{
		hmac:     hmac,
		endpoint: endpoint,
		service:  s,
		hc:       utils.HttpClient{MaxRetries: 3},
	}
}

type robot struct {
	hmac     string
	endpoint string
	hc       utils.HttpClient
	service  sync.SyncService
}

func (bot *robot) HandlePushEvent(e *sdk.PushEvent, log *logrus.Entry) (err error) {
	repoName := e.Project.Name

	owner, err := domain.NewAccount(e.Project.Namespace)
	if err != nil {
		return
	}

	v := sync.RepoInfo{
		Owner:    owner,
		RepoId:   strconv.Itoa(e.ProjectID),
		RepoName: repoName,
	}

	if err = bot.service.SyncRepo(&v); err == nil {
		return
	}

	if err1 := bot.sendBack(e); err1 != nil {
		log.Errorf(
			"sync repo failed and send back event failed, err:%s.",
			err1.Error(),
		)
	}

	return
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
