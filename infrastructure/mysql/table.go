package mysql

import "strconv"

const (
	fieldStatus     = "status"
	fieldVersion    = "version"
	fieldLastCommit = "last_commit"
)

var (
	modelTableName   = ""
	datasetTableName = ""
)

type repoSyncLock interface {
	GetId() string
}

type RepoSyncLock struct {
	Id         int    `json:"-"            gorm:"column:id"`
	Owner      string `json:"-"            gorm:"column:owner"`
	RepoId     string `json:"-"            gorm:"column:repo_id"`
	Status     string `json:"status"       gorm:"column:status"`
	Version    int    `json:"-"            gorm:"column:version"`
	LastCommit string `json:"last_commit"  gorm:"column:last_commit"`
}

type ModelRepoSyncLock struct {
	*RepoSyncLock `gorm:"embedded"`
}

func (r *ModelRepoSyncLock) TableName() string {
	return modelTableName
}

func (r *ModelRepoSyncLock) GetId() string {
	return strconv.Itoa(r.Id)
}

type DatasetRepoSyncLock struct {
	*RepoSyncLock `gorm:"embedded"`
}

func (r *DatasetRepoSyncLock) TableName() string {
	return datasetTableName
}

func (r *DatasetRepoSyncLock) GetId() string {
	return strconv.Itoa(r.Id)
}
