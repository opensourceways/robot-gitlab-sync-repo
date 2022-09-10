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

func (s *syncCommit) repoOBSPath() string {
	return filepath.Join(s.owner, s.repoType, s.repoId)
}

type syncService struct {
	obs           *syncToOBS
	workDir       string
	isFirstCommit func(string, string) (bool, error)
	commitFileSh  string
}

func (s syncService) sync(commit syncCommit) error {
	v, unavailable, err := s.obs.getCurrentCommit(commit.repoOBSPath())
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

	return s.obs.updateCurrentCommit(commit.repoOBSPath(), commit.commit)
}

func (s syncService) doSync(commit *syncCommit) error {
	tempDir, err := ioutil.TempDir(s.workDir, "sync")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	smallFile, lfsFile, err := s.getCommitFile(tempDir, commit)
	if err != nil {
		return err
	}

	if smallFile != "" {
		err := s.syncSmallFiles(
			smallFile, filepath.Join(tempDir, commit.repoName),
			commit.repoOBSPath(),
		)
		if err != nil {
			return err
		}
	}

	if lfsFile != "" {
		err := s.syncLFSFiles(lfsFile, commit.repoOBSPath())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s syncService) syncSmallFiles(smallFiles, repoDir, obsPath string) error {
	return utils.ReadFileLineByLine(smallFiles, func(line string) bool {
		src := filepath.Join(repoDir, line)
		dst := filepath.Join(obsPath, line)

		if err := s.obs.syncSmallFile(src, dst); err != nil {
			return true
		}

		return false
	})
}

func (s syncService) syncLFSFiles(lfsFiles, obsPath string) error {
	return utils.ReadFileLineByLine(lfsFiles, func(line string) bool {
		v := strings.Split(line, ":oid sha256:")
		dst := filepath.Join(obsPath, v[0])

		if err := s.obs.syncLFSFile(v[1], dst); err != nil {
			return true
		}

		return false
	})
}

func (s syncService) getCommitFile(workDir string, commit *syncCommit) (
	smallFile string, lfsFile string, err error,
) {
	v, err, _ := utils.RunCmd(
		s.commitFileSh, workDir, commit.repoURL,
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
