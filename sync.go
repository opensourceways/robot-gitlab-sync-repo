package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type syncToOBS struct {
	obsClient  *obs.ObsClient
	bucketName string
	lfsDir     string
}

// targetDir: xihe-obj/user/[project,model,dataset]/repo_id
func (s syncToOBS) syncSmallFiles(repoDir string, files []string, targetDir string) error {
	// TODO need do concurrently?

	for _, f := range files {
		err := s.retry(func() error {
			return s.createOBSObject(
				filepath.Join(repoDir, f),
				filepath.Join(targetDir, f),
			)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// targetDir: xihe-obj/user/[project,model,dataset]/repo_id
func (s syncToOBS) syncLFSFiles(fileSha map[string]string, targetDir string) error {
	// TODO need do concurrently?

	for f, sha := range fileSha {
		err := s.retry(func() error {
			return s.copyOBSObject(sha, filepath.Join(targetDir, f))
		})

		if err != nil {
			return err
		}

	}

	return nil
}

func (s syncToOBS) retry(f func() error) (err error) {
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

func (s syncToOBS) createOBSObject(from, to string) error {
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

func (s syncToOBS) copyOBSObject(sha, to string) error {
	input := &obs.CopyObjectInput{}
	input.Bucket = s.bucketName
	input.Key = to
	input.CopySourceBucket = s.bucketName
	input.CopySourceKey = filepath.Join(s.lfsDir, sha[:2], sha[2:4], sha[4:])

	_, err := s.obsClient.CopyObject(input)

	return err
}
