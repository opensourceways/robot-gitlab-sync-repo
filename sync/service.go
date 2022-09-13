package sync

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opensourceways/robot-gitlab-sync-repo/domain"
	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
)

type SyncCommit struct {
	Owner        string
	RepoId       int
	RepoURL      string
	RepoType     string
	RepoName     string
	Commit       string
	ParentCommit string
}

func (s *SyncCommit) repoOBSPath() string {
	return filepath.Join(s.Owner, s.RepoType, strconv.Itoa(s.RepoId))
}

type syncService struct {
	obs           *syncToOBS
	workDir       string
	commitFileSh  string
	obsutil       string
	syncRepo      domain.Repository
	isFirstCommit func(string, string) (bool, error)
	getLastCommit func(string, int) (string, error)
}

func (s *syncService) SyncRepo(commit *SyncCommit) error {
	// if 404, create in the Find
	c, err := s.syncRepo.Find(commit.Owner, commit.RepoName)
	if err != nil {
		return err
	}

	if c.Status != "done" {
		return errors.New("can't sync")
	}

	lastCommit, err := s.getLastCommit(commit.Owner, commit.RepoId)
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
	lastCommit, err = s.sync(commit)

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

func (s *syncService) sync(commit *SyncCommit) (last string, err error) {
	tempDir, err := ioutil.TempDir(s.workDir, "sync")
	if err != nil {
		return
	}

	defer os.RemoveAll(tempDir)

	last, lfsFile, err := s.getCommitFile(tempDir, commit)
	if err != nil {
		return
	}

	if lfsFile != "" {
		if err = s.syncLFSFiles(lfsFile, commit.repoOBSPath()); err != nil {
			return
		}
	}

	err = s.obs.updateCurrentCommit(commit.repoOBSPath(), last)

	return
}

func (s *syncService) syncLFSFiles(lfsFiles, obsPath string) error {
	return utils.ReadFileLineByLine(lfsFiles, func(line string) bool {
		v := strings.Split(line, ":oid sha256:")
		dst := filepath.Join(obsPath, v[0])

		if err := s.obs.syncLFSFile(v[1], dst); err != nil {
			return true
		}

		return false
	})
}

func (s *syncService) getCommitFile(workDir string, commit *SyncCommit) (
	lastCommit string, lfsFile string, err error,
) {
	c, err := s.obs.getCurrentCommit(commit.repoOBSPath())
	if err != nil {
		return
	}

	obspath := s.obs.getRepoObsPath(commit.repoOBSPath())

	v, err, _ := utils.RunCmd(
		s.commitFileSh, workDir, commit.RepoURL,
		commit.RepoName, c, s.obsutil, obspath,
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
