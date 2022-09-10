package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
)

type syncToOBS struct {
	obsClient         *obs.ObsClient
	lfsPath           string
	repoPath          string
	bucketName        string
	currentCommitFile string // config
}

// src: local file
// dst: user/[project,model,dataset]/repo_id/xxx
func (s *syncToOBS) syncSmallFile(src, dst string) error {
	return utils.Retry(func() error {
		return s.uploadFileToOBS(src, filepath.Join(s.repoPath, dst))
	})
}

// sha: sha
// dst: user/[project,model,dataset]/repo_id/xxx
func (s *syncToOBS) syncLFSFile(sha, dst string) error {
	return utils.Retry(func() error {
		return s.copyOBSObject(
			filepath.Join(s.lfsPath, sha[:2], sha[2:4], sha[4:]),
			filepath.Join(s.repoPath, dst))
	})
}

// p: user/[project,model,dataset]/repo_id
func (s *syncToOBS) getCurrentCommit(p string) (d []byte, unavailable bool, err error) {
	utils.Retry(func() error {
		d, unavailable, err = s.getOBSObject(
			filepath.Join(s.repoPath, p, s.currentCommitFile),
		)

		return err
	})

	return
}

// p: user/[project,model,dataset]/repo_id
func (s *syncToOBS) updateCurrentCommit(p, commit string) error {
	return utils.Retry(func() error {
		return s.saveToOBS(
			filepath.Join(s.repoPath, p, s.currentCommitFile),
			commit,
		)
	})
}

func (s *syncToOBS) uploadFileToOBS(from, to string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}

	defer f.Close()

	md5, err := utils.GenMd5OfByteStream(f)
	if err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	input := &obs.PutObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.Body = f
	input.ContentMD5 = md5

	_, err = s.obsClient.PutObject(input)

	return err
}

func (s *syncToOBS) saveToOBS(to, content string) error {
	input := &obs.PutObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.Body = strings.NewReader(content)
	input.ContentMD5 = utils.GenMD5([]byte(content))

	_, err := s.obsClient.PutObject(input)

	return err
}

func (s *syncToOBS) copyOBSObject(from, to string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.CopySourceBucket = s.bucketName
	input.CopySourceKey = from

	_, err := s.obsClient.CopyObject(input)

	return err
}

func (s *syncToOBS) getOBSObject(p string) ([]byte, bool, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = s.bucketName
	input.Key = p

	output, err := s.obsClient.GetObject(input)
	if err != nil {
		v, ok := err.(obs.ObsError)
		if ok && v.BaseModel.StatusCode == 404 {
			return nil, true, nil
		}

		return nil, false, err
	}

	defer output.Body.Close()

	v, err := ioutil.ReadAll(output.Body)

	return v, false, err
}