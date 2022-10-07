package sync

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opensourceways/robot-gitlab-sync-repo/domain/obs"
	"github.com/opensourceways/robot-gitlab-sync-repo/domain/platform"
	"github.com/opensourceways/robot-gitlab-sync-repo/domain/repository"
	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
)

type RepoInfo struct {
	Owner    string
	RepoId   string
	RepoType string
	RepoName string
}

func (s *RepoInfo) repoOBSPath() string {
	return filepath.Join(s.Owner, s.RepoType, s.RepoId)
}

type SyncService interface {
	SyncRepo(*RepoInfo) error
}

func NewSyncService(
	cfg *Config, s obs.OBS, syncRepo repository.RepoSync,
	p platform.Platform,
) SyncService {
	return &syncService{
		h: &syncHelper{
			obsService:        s,
			lfsPath:           cfg.LFSPath,
			repoPath:          cfg.RepoPath,
			currentCommitFile: cfg.CommitFile,
		},
		workDir:    cfg.WorkDir,
		obsutil:    s.OBSUtilPath(),
		syncFileSh: cfg.SyncFileShell,
		syncRepo:   syncRepo,
		ph:         p,
	}
}

type syncService struct {
	h          *syncHelper
	workDir    string
	obsutil    string
	syncFileSh string
	syncRepo   repository.RepoSync
	ph         platform.Platform
}

func (s *syncService) SyncRepo(info *RepoInfo) error {
	// if 404, create in the Find
	c, err := s.syncRepo.Find(info.Owner, info.RepoId)
	if err != nil {
		return err
	}

	if c.Status != "" && c.Status != "done" {
		return errors.New("can't sync")
	}

	lastCommit, err := s.ph.GetLastCommit(info.RepoId)
	if err != nil {
		return err
	}

	if c.LastCommit == lastCommit {
		return nil
	}

	c.Status = "running"
	c, err = s.syncRepo.Save(&c)
	if err != nil {
		return err
	}

	// do sync
	lastCommit, err = s.sync(info)

	// update
	c.Status = "done"
	if err == nil {
		c.LastCommit = lastCommit
	}

	err1 := utils.Retry(func() error {
		if _, err := s.syncRepo.Save(&c); err != nil {
			// log
		}

		return nil
	})
	if err1 != nil {
		// dead lock happend for this repo
	}

	return err
}

func (s *syncService) sync(info *RepoInfo) (last string, err error) {
	tempDir, err := ioutil.TempDir(s.workDir, "sync")
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
		s.syncFileSh, workDir,
		s.ph.GetCloneURL(info.Owner, info.RepoName),
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
