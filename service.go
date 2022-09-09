package main

import (
	"errors"
	"os"
	"path/filepath"
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
	// TODO gen tmp dir
	dir := ""

	defer os.RemoveAll(dir)

	smalls, lfsSHA, err := s.getCommitFile(dir, commit)
	if err != nil {
		return err
	}

	if len(smalls) > 0 {
		err := s.obs.syncSmallFiles(
			filepath.Join(dir, commit.repoName),
			smalls, commit.repoPath(),
		)

		if err != nil {
			return err
		}
	}

	if len(lfsSHA) > 0 {
		err := s.obs.syncLFSFiles(lfsSHA, commit.repoPath())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s syncService) getCommitFile(dir string, commit *syncCommit) (
	smalls []string, lfsSHA map[string]string, err error,
) {
	/*
		work_dir=$1
		repo_url=$2
		repo_name=$3
		commit=$4
	*/

	return
}
