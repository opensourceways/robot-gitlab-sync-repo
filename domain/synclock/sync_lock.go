package synclock

import (
	"github.com/opensourceways/robot-gitlab-sync-repo/domain"
)

type ErrorDuplicateCreating struct {
	error
}

func NewErrorDuplicateCreating(err error) ErrorDuplicateCreating {
	return ErrorDuplicateCreating{err}
}

type ErrorRepoNotExists struct {
	error
}

func NewErrorRepoNotExists(err error) ErrorRepoNotExists {
	return ErrorRepoNotExists{err}
}

type ErrorConcurrentUpdating struct {
	error
}

func NewErrorConcurrentUpdating(err error) ErrorConcurrentUpdating {
	return ErrorConcurrentUpdating{err}
}

func IsRepoSyncLockNotExist(err error) bool {
	_, ok := err.(ErrorRepoNotExists)

	return ok
}

type RepoSyncLock interface {
	Find(
		owner domain.Account,
		repoType domain.ResourceType, repoId string,
	) (domain.RepoSyncLock, error)

	Save(*domain.RepoSyncLock) (domain.RepoSyncLock, error)
}
