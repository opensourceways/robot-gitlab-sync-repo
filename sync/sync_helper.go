package sync

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
)

type syncHelper struct {
	obsClient         *obs.ObsClient
	lfsPath           string
	repoPath          string
	bucketName        string
	currentCommitFile string // config
}

// sha: sha
// dst: user/[project,model,dataset]/repo_id/xxx
func (s *syncHelper) syncLFSFile(sha, dst string) error {
	return utils.Retry(func() error {
		return s.copyOBSObject(
			filepath.Join(s.lfsPath, sha[:2], sha[2:4], sha[4:]),
			filepath.Join(s.repoPath, dst))
	})
}

// p: user/[project,model,dataset]/repo_id
func (s *syncHelper) getCurrentCommit(p string) (c string, err error) {
	err = utils.Retry(func() error {
		v, err := s.getOBSObject(
			filepath.Join(s.repoPath, p, s.currentCommitFile),
		)
		if err == nil && len(v) > 0 {
			c = string(v)
		}

		return err
	})

	return
}

// p: user/[project,model,dataset]/repo_id
func (s *syncHelper) updateCurrentCommit(p, commit string) error {
	return utils.Retry(func() error {
		return s.saveToOBS(
			filepath.Join(s.repoPath, p, s.currentCommitFile),
			commit,
		)
	})
}

// p: user/[project,model,dataset]/repo_id
func (s *syncHelper) getRepoObsPath(p string) string {
	return filepath.Join(s.repoPath, p)
}

func (s *syncHelper) saveToOBS(to, content string) error {
	input := &obs.PutObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.Body = strings.NewReader(content)
	input.ContentMD5 = utils.GenMD5([]byte(content))

	_, err := s.obsClient.PutObject(input)

	return err
}

func (s *syncHelper) copyOBSObject(from, to string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.CopySourceBucket = s.bucketName
	input.CopySourceKey = from

	_, err := s.obsClient.CopyObject(input)

	return err
}

func (s *syncHelper) getOBSObject(p string) ([]byte, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = s.bucketName
	input.Key = p

	output, err := s.obsClient.GetObject(input)
	if err != nil {
		v, ok := err.(obs.ObsError)
		if ok && v.BaseModel.StatusCode == 404 {
			return nil, nil
		}

		return nil, err
	}

	defer output.Body.Close()

	return ioutil.ReadAll(output.Body)
}
