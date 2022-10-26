package mysql

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/opensourceways/robot-gitlab-sync-repo/domain"
	"github.com/opensourceways/robot-gitlab-sync-repo/infrastructure/synclockimpl"
)

func NewSyncLockMapper() synclockimpl.SyncLockMapper {
	return syncLock{}
}

type syncLock struct{}

func (rs syncLock) getTable(t string, v *RepoSyncLock) interface{} {
	switch t {
	case domain.ResourceTypeModel.ResourceType():
		return ModelRepoSyncLock{v}

	case domain.ResourceTypeDataset.ResourceType():
		return DatasetRepoSyncLock{v}

	default:
		return nil
	}
}

func (rs syncLock) Insert(do *synclockimpl.RepoSyncLockDO) (string, error) {
	data := rs.toSyncLockTable(do)
	table := rs.getTable(do.RepoType, &data)

	r := cli.db.Model(table).Create(table)
	if r.Error != nil {
		return "", r.Error
	}

	if r.RowsAffected == 0 {
		return "", synclockimpl.NewErrorDuplicateCreating(
			errors.New("duplecate creating"),
		)
	}

	return table.(repoSyncLock).GetId(), nil
}

func (rs syncLock) Get(owner, repoType, repoId string) (do synclockimpl.RepoSyncLockDO, err error) {
	cond := &RepoSyncLock{
		Owner:  owner,
		RepoId: repoId,
	}

	data := new(RepoSyncLock)
	table := rs.getTable(repoType, nil)

	err = cli.db.Model(table).Where(cond).First(data).Error

	if err == nil {
		do = rs.toSyncLockDo(data, repoType)
	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = synclockimpl.NewErrorDataNotExists(err)
		}
	}

	return
}

func (rs syncLock) Update(do *synclockimpl.RepoSyncLockDO) error {
	cond := &RepoSyncLock{
		Owner:   do.Owner,
		RepoId:  do.RepoId,
		Version: do.Version,
	}

	table := rs.getTable(do.RepoType, nil)

	tx := cli.db.Model(table).Where(cond).Updates(
		map[string]interface{}{
			fieldVersion:    gorm.Expr("? + ?", fieldVersion, 1),
			fieldLastCommit: do.LastCommit,
			fieldStatus:     do.Status,
		},
	)
	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		return synclockimpl.NewErrorConcurrentUpdating(
			errors.New("no matched record"),
		)
	}

	return nil
}

func (rs syncLock) toSyncLockTable(do *synclockimpl.RepoSyncLockDO) RepoSyncLock {
	return RepoSyncLock{
		Owner:      do.Owner,
		RepoId:     do.RepoId,
		Status:     do.Status,
		Version:    do.Version,
		LastCommit: do.LastCommit,
	}
}

func (rs syncLock) toSyncLockDo(data *RepoSyncLock, repoType string) synclockimpl.RepoSyncLockDO {
	return synclockimpl.RepoSyncLockDO{
		Id:         strconv.Itoa(data.Id),
		Owner:      data.Owner,
		RepoId:     data.RepoId,
		Status:     data.Status,
		Version:    data.Version,
		RepoType:   repoType,
		LastCommit: data.LastCommit,
	}
}
