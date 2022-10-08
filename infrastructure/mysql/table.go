package mysql

import "github.com/opensourceways/robot-gitlab-sync-repo/domain"

const (
	fieldStatus     = "status"
	fieldVersion    = "version"
	fieldLastCommit = "last_commit"
)

type repoSyncLock interface {
	repoSyncLock() *RepoSyncLock
}

// TODO verify if it needs define json
type RepoSyncLock struct {
	Id         int    `json:"id"           gorm:"column:id"`
	Owner      string `json:"owner"        gorm:"column:owner"`
	RepoId     string `json:"repo_id"      gorm:"column:repo_id"`
	Status     string `json:"status"       gorm:"column:status"`
	Version    int    `json:"version"      gorm:"column:version"`
	LastCommit string `json:"last_commit"  gorm:"column:last_commit"`
}

type ProjectRepoSyncLock struct {
	RepoSyncLock `gorm:"embedded"`
}

func (r *ProjectRepoSyncLock) TableName() string {
	return domain.ResourceTypeProject.ResourceType()
}

func (r *ProjectRepoSyncLock) repoSyncLock() *RepoSyncLock {
	return &r.RepoSyncLock
}

type ModelRepoSyncLock struct {
	RepoSyncLock `gorm:"embedded"`
}

func (r *ModelRepoSyncLock) TableName() string {
	return domain.ResourceTypeModel.ResourceType()
}

func (r *ModelRepoSyncLock) repoSyncLock() *RepoSyncLock {
	return &r.RepoSyncLock
}

type DatasetRepoSyncLock struct {
	RepoSyncLock `gorm:"embedded"`
}

func (r *DatasetRepoSyncLock) TableName() string {
	return domain.ResourceTypeDataset.ResourceType()
}

func (r *DatasetRepoSyncLock) repoSyncLock() *RepoSyncLock {
	return &r.RepoSyncLock
}
