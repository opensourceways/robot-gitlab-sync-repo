package utils

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"time"
)

func GenMD5(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}

func GenMd5OfByteStream(r io.Reader) (string, error) {
	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func ReadFileLineByLine(filename string, handle func(string) bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if b := handle(scanner.Text()); b {
			break
		}
	}

	return nil
}

func Retry(f func() error) (err error) {
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
