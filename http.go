package goutils

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"time"
)

// Gauger 定义了用于跟踪下载进度的接口。
type Gauger interface {
	Progress(current, total uint64, completed bool)
}

type gaugeReader struct {
	io.Reader

	total      uint64
	current    uint64
	lastUpdate int64
	gaugers    []Gauger
}

func (gr *gaugeReader) Read(p []byte) (n int, err error) {
	n, err = gr.Reader.Read(p)
	gr.current += uint64(n)
	if gr.total > 0 && gr.current >= gr.total {
		for _, gauger := range gr.gaugers {
			gauger.Progress(gr.total, gr.total, false)
		}

		return
	}

	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-gr.lastUpdate > 100 {
		for _, gauger := range gr.gaugers {
			gauger.Progress(gr.current, gr.total, false)
		}
		gr.lastUpdate = now
	}

	return
}

func (gr *gaugeReader) completed() {
	for _, gauger := range gr.gaugers {
		gauger.Progress(gr.current, gr.total, true)
	}
}

func DownloadFile(url, dest string, validateCert bool) (err error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
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

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)

	return err
}

func DownloadFileWithGauger(url, dest string, validateCert bool, gaugers ...Gauger) (err error) {
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

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var total uint64
	if resp.ContentLength > 0 {
		total = uint64(resp.ContentLength)
	}
	var reader io.Reader = resp.Body
	gr := &gaugeReader{
		Reader:  reader,
		total:   total,
		gaugers: gaugers,
	}
	if len(gaugers) > 0 {
		reader = gr
	}

	// Write the body to file
	_, err = io.Copy(out, reader)
	if err == nil && len(gaugers) > 0 {
		gr.completed()
	}

	return err
}
