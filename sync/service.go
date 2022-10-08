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
	syncRepo synclock.RepoSyncLock,
	p platform.Platform,
) SyncService {
	return &syncService{
		h: &syncHelper{
			obsService: s,
			cfg:        cfg.HelperConfig,
		},
		log:      log,
		cfg:      cfg.ServiceConfig,
		obsutil:  s.OBSUtilPath(),
		syncRepo: syncRepo,
		ph:       p,
	}
}

type syncService struct {
	h       *syncHelper
	log     *logrus.Entry
	cfg     ServiceConfig
	obsutil string

	syncRepo synclock.RepoSyncLock
	ph       platform.Platform
}

func (s *syncService) SyncRepo(info *RepoInfo) error {
	c, err := s.syncRepo.Find(info.Owner, info.RepoType, info.RepoId)
	if err != nil && !synclock.IsRepoSyncLockNotExist(err) {
		return err
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
	c, err = s.syncRepo.Save(&c)
	if err != nil {
		return err
	}

	// do sync
	lastCommit, err = s.sync(info)

	// update
	c.Status = domain.RepoSyncStatusDone
	if err == nil {
		c.LastCommit = lastCommit
	}

	err1 := utils.Retry(func() error {
		if _, err := s.syncRepo.Save(&c); err != nil {
			s.log.Errorf(
				"save sync repo(%s) failed, err:%s",
				info.repoOBSPath, err.Error(),
			)
		}

		return nil
	})
	if err1 != nil {
		s.log.Errorf(
			"save sync repo(%s) failed, dead lock happened",
			info.repoOBSPath,
		)
	}

	return err
}

func (s *syncService) sync(info *RepoInfo) (last string, err error) {
	tempDir, err := ioutil.TempDir(s.cfg.WorkDir, "sync")
	if err != nil {
		return
	}

	defer os.RemoveAll(tempDir)

	last, lfsFile, err := s.syncFile(tempDir, info)
	if err != nil {
		return
	}

	if lfsFile != "" {
		if err = s.syncLFSFiles(lfsFile, info); err != nil {
			return
		}
	}

	err = s.h.updateCurrentCommit(info.repoOBSPath(), last)

	return
}

func (s *syncService) syncLFSFiles(lfsFiles string, info *RepoInfo) error {
	obsPath := info.repoOBSPath()

	return utils.ReadFileLineByLine(lfsFiles, func(line string) bool {
		v := strings.Split(line, ":oid sha256:")
		dst := filepath.Join(obsPath, v[0])

		if err := s.h.syncLFSFile(v[1], dst); err != nil {
			return true
		}

		return false
	})
}

func (s *syncService) syncFile(workDir string, info *RepoInfo) (
	lastCommit string, lfsFile string, err error,
) {
	p := info.repoOBSPath()
	c, err := s.h.getCurrentCommit(p)
	if err != nil {
		return
	}

	obspath := s.h.getRepoObsPath(p)
	if !strings.HasPrefix(obspath, "/") {
		obspath += "/"
	}

	v, err, _ := utils.RunCmd(
		s.cfg.SyncFileShell, workDir,
		s.ph.GetCloneURL(info.Owner.Account(), info.RepoName),
		info.RepoName, c, s.obsutil, obspath,
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
