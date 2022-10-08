package sync

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opensourceways/robot-gitlab-sync-repo/domain"
	"github.com/opensourceways/robot-gitlab-sync-repo/domain/obs"
	"github.com/opensourceways/robot-gitlab-sync-repo/domain/platform"
	"github.com/opensourceways/robot-gitlab-sync-repo/domain/synclock"
	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
	"github.com/sirupsen/logrus"
)

type RepoInfo struct {
	Owner    domain.Account
	RepoId   string
	RepoType domain.ResourceType
	RepoName string
}

func (s *RepoInfo) repoOBSPath() string {
	return filepath.Join(
		s.Owner.Account(), s.RepoType.ResourceType(),
		s.RepoId,
	)
}

type SyncService interface {
	SyncRepo(*RepoInfo) error
}

func NewSyncService(
	cfg *Config, log *logrus.Entry,
	s obs.OBS,
	p platform.Platform,
	l synclock.RepoSyncLock,
) SyncService {
	return &syncService{
		h: &syncHelper{
			obsService: s,
			cfg:        cfg.HelperConfig,
		},
		log:     log,
		cfg:     cfg.ServiceConfig,
		obsutil: s.OBSUtilPath(),
		lock:    l,
		ph:      p,
	}
}

type syncService struct {
	h       *syncHelper
	log     *logrus.Entry
	cfg     ServiceConfig
	obsutil string

	lock synclock.RepoSyncLock
	ph   platform.Platform
}

func (s *syncService) SyncRepo(info *RepoInfo) error {
	c, err := s.lock.Find(info.Owner, info.RepoType, info.RepoId)
	if err != nil {
		if !synclock.IsRepoSyncLockNotExist(err) {
			return err
		}

		c.Owner = info.Owner
		c.RepoId = info.RepoId
		c.RepoType = info.RepoType
	}

	if c.Status != nil && !c.Status.IsDone() {
		return errors.New("can't sync")
	}

	lastCommit, err := s.ph.GetLastCommit(info.RepoId)
	if err != nil {
		return err
	}

	if c.LastCommit == lastCommit {
		return nil
	}

	c.Status = domain.RepoSyncStatusRunning
	c, err = s.lock.Save(&c)
	if err != nil {
		return err
	}

	// do sync
	lastCommit, syncErr := s.sync(c.LastCommit, info)
	if syncErr == nil {
		c.LastCommit = lastCommit

		err := s.h.saveLastCommit(info.repoOBSPath(), lastCommit)
		if err != nil {
			s.log.Errorf(
				"update last commit failed, err:%s",
				err.Error(),
			)
		}
	}
	c.Status = domain.RepoSyncStatusDone

	err = utils.Retry(func() error {
		if _, err := s.lock.Save(&c); err != nil {
			s.log.Errorf(
				"save sync repo(%s) failed, err:%s",
				info.repoOBSPath, err.Error(),
			)
		}

		return nil
	})
	if err != nil {
		s.log.Errorf(
			"save sync repo(%s) failed, dead lock happened",
			info.repoOBSPath,
		)
	}

	return syncErr
}

func (s *syncService) sync(startCommit string, info *RepoInfo) (last string, err error) {
	tempDir, err := ioutil.TempDir(s.cfg.WorkDir, "sync")
	if err != nil {
		return
	}

	defer os.RemoveAll(tempDir)

	last, lfsFile, err := s.syncFile(tempDir, startCommit, info)
	if err != nil || lfsFile == "" {
		return
	}

	err = s.syncLFSFiles(lfsFile, info)

	return
}

func (s *syncService) syncLFSFiles(lfsFiles string, info *RepoInfo) error {
	obsPath := info.repoOBSPath()

	return utils.ReadFileLineByLine(lfsFiles, func(line string) error {
		v := strings.Split(line, ":oid sha256:")
		dst := filepath.Join(obsPath, v[0])

		return s.h.syncLFSFile(v[1], dst)
	})
}

func (s *syncService) syncFile(workDir, startCommit string, info *RepoInfo) (
	lastCommit string, lfsFile string, err error,
) {
	obspath := s.h.getRepoObsPath(info.repoOBSPath())
	if !strings.HasPrefix(obspath, "/") {
		obspath += "/"
	}

	v, err, _ := utils.RunCmd(
		s.cfg.SyncFileShell, workDir,
		s.ph.GetCloneURL(info.Owner.Account(), info.RepoName),
		info.RepoName, startCommit, s.obsutil, obspath,
	)
	if err != nil {
		return
	}

	r := strings.Split(string(v), ", ")
	lastCommit = r[0]

	if r[2] == "yes" {
		lfsFile = r[1]
	}

	return
}
