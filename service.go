package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
)

type syncCommit struct {
	owner        string
	repoType     string
	repoURL      string
	repoName     string
	repoId       string
	commit       string
	parentCommit string
}

func (s syncCommit) repoPath() string {
	return filepath.Join(s.owner, s.repoType, s.repoId)
}

type syncService struct {
	obs           *syncToOBS
	workDir       string
	isFirstCommit func(string, string) (bool, error)
	commitFileSh  string
}

func (s syncService) sync(commit syncCommit) error {
	v, unavailable, err := s.obs.getCurrentCommit(commit.repoPath())
	if err != nil {
		return err
	}

	if unavailable {
		b, err := s.isFirstCommit(commit.parentCommit, commit.commit)
		if err != nil {
			return err
		}

		if !b {
			return errors.New("can't sync commit")
		}

	} else if string(v) != commit.parentCommit {
		return errors.New("can't sync commit")
	}

	// can sync
	if err := s.doSync(&commit); err != nil {
		return err
	}

	return s.obs.updateCurrentCommit(commit.repoPath(), commit.commit)
}

func (s syncService) doSync(commit *syncCommit) error {
	dir, err := ioutil.TempDir(s.workDir, "sync")
	if err != nil {
		return err
	}

	defer os.RemoveAll(dir)

	smallFile, lfsFile, err := s.getCommitFile(dir, commit)
	if err != nil {
		return err
	}

	if smallFile != "" {
		err := s.syncSmallFiles(
			smallFile, filepath.Join(dir, commit.repoName),
			commit.repoPath(),
		)

		if err != nil {
			return err
		}
	}

	if lfsFile != "" {
		err := s.syncLFSFiles(lfsFile, commit.repoPath())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s syncService) syncSmallFiles(file, repoDir, obsPath string) error {
	return utils.ReadFileLineByLine(file, func(line string) bool {
		if err := s.obs.syncSmallFile(repoDir, file, obsPath); err != nil {
			return true
		}

		return false
	})
}

func (s syncService) syncLFSFiles(file, obsPath string) error {
	return utils.ReadFileLineByLine(file, func(line string) bool {
		v := strings.Split(line, ":oid sha256:")
		if err := s.obs.syncLFSFile(v[0], v[1], obsPath); err != nil {
			return true
		}

		return false
	})
}

func (s syncService) getCommitFile(dir string, commit *syncCommit) (
	smallFile string, lfsFile string, err error,
) {
	v, err, _ := utils.RunCmd(
		s.commitFileSh, dir, commit.repoURL,
		commit.repoName, commit.commit,
	)

	if err != nil {
		return
	}

	r := strings.Split(string(v), ", ")

	if r[1] == "yes" {
		smallFile = r[0]
	}

	if r[3] == "yes" {
		lfsFile = r[2]
	}

	return
}
