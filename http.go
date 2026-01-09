package goutils

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"time"
)

type Progresses interface {
	Progress(current, total uint64)
}

type progressReader struct {
	io.Reader
	total      uint64
	current    uint64
	lastUpdate int64
	progresses []Progresses
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.current += uint64(n)
	if pr.current >= pr.total {
		for _, progress := range pr.progresses {
			progress.Progress(pr.total, pr.total)
		}

		return
	}

	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-pr.lastUpdate > 100 {
		for _, progress := range pr.progresses {
			progress.Progress(pr.current, pr.total)
		}
		pr.lastUpdate = now
	}

	return
}

func DownloadFile(url, dest string, validateCert bool, progresses ...Progresses) (err error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !validateCert,
			},
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	totalSize := resp.ContentLength

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	if totalSize > 0 && len(progresses) > 0 {
		reader = &progressReader{
			Reader:     reader,
			total:      uint64(totalSize),
			progresses: progresses,
		}
	}

	// Write the body to file
	_, err = io.Copy(out, reader)

	return err
}
