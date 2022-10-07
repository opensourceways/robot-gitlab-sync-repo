package repository

import "github.com/opensourceways/robot-gitlab-sync-repo/domain"

type RepoSync interface {
	Find(owner, repoId string) (domain.RepoSync, error)
	Save(*domain.RepoSync) (domain.RepoSync, error)
}
