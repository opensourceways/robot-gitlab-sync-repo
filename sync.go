package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

	"github.com/opensourceways/robot-gitlab-sync-repo/utils"
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
			return s.uploadFileToOBS(
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
func (s *syncToOBS) syncSmallFile(repoDir string, file string, targetDir string) error {
	return s.retry(func() error {
		return s.uploadFileToOBS(
			filepath.Join(repoDir, file),
			filepath.Join(s.repoDir, targetDir, file),
		)
	})
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

// targetDir: user/[project,model,dataset]/repo_id
func (s *syncToOBS) syncLFSFile(file string, sha string, targetDir string) error {
	return s.retry(func() error {
		return s.copyOBSObject(sha, filepath.Join(s.repoDir, targetDir, file))
	})
}

// p: user/[project,model,dataset]/repo_id
func (s *syncToOBS) getCurrentCommit(p string) (d []byte, unavailable bool, err error) {
	s.retry(func() error {
		d, unavailable, err = s.getOBSObject(
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
