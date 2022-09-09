package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type syncToOBS struct {
	obsClient         *obs.ObsClient
	bucketName        string
	lfsDir            string
	repoDir           string
	currentCommitFile string // config
}

// targetDir: user/[project,model,dataset]/repo_id
func (s *syncToOBS) syncSmallFiles(repoDir string, files []string, targetDir string) error {
	// TODO need do concurrently?

	for _, f := range files {
		err := s.retry(func() error {
			return s.createOBSObject(
				filepath.Join(repoDir, f),
				filepath.Join(s.repoDir, targetDir, f),
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// targetDir: user/[project,model,dataset]/repo_id
func (s *syncToOBS) syncLFSFiles(fileSha map[string]string, targetDir string) error {
	// TODO need do concurrently?

	for f, sha := range fileSha {
		err := s.retry(func() error {
			return s.copyOBSObject(sha, filepath.Join(s.repoDir, targetDir, f))
		})

		if err != nil {
			return err
		}

	}

	return nil
}

// p: user/[project,model,dataset]/repo_id
func (s *syncToOBS) getCurrentCommit(p string) (d []byte, available bool, err error) {
	s.retry(func() error {
		d, available, err = s.getOBSObject(
			filepath.Join(s.repoDir, p, s.currentCommitFile),
		)

		return err
	})

	return
}

// p: user/[project,model,dataset]/repo_id
func (s *syncToOBS) updateCurrentCommit(p, commit string) error {
	return s.retry(func() error {
		return s.saveToOBS(
			filepath.Join(s.repoDir, p, s.currentCommitFile),
			commit,
		)
	})
}

func (s *syncToOBS) retry(f func() error) (err error) {
	if err = f(); err == nil {
		return
	}

	t := 100 * time.Millisecond

	for i := 1; i < 10; i++ {
		time.Sleep(t)

		if err = f(); err == nil {
			return
		}
	}

	return
}

func (s *syncToOBS) createOBSObject(from, to string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}

	defer f.Close()

	input := &obs.PutObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.Body = f

	_, err = s.obsClient.PutObject(input)

	return err
}

func (s *syncToOBS) saveToOBS(to, content string) error {
	input := &obs.PutObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.Body = strings.NewReader(content)

	_, err := s.obsClient.PutObject(input)

	return err
}

func (s *syncToOBS) copyOBSObject(sha, to string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.CopySourceBucket = s.bucketName
	input.CopySourceKey = filepath.Join(s.lfsDir, sha[:2], sha[2:4], sha[4:])

	_, err := s.obsClient.CopyObject(input)

	return err
}

func (s *syncToOBS) getOBSObject(p string) ([]byte, bool, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = s.bucketName
	input.Key = p

	output, err := s.obsClient.GetObject(input)
	if err != nil {
		if v, ok := err.(obs.ObsError); ok && v.BaseModel.StatusCode == 404 {
			return nil, true, nil
		}

		return nil, false, err
	}

	defer output.Body.Close()

	v, err := ioutil.ReadAll(output.Body)

	return v, false, err
}
