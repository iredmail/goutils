package goutils

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"time"
)

func DownloadFile(url, dest string, validateCert bool) (err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !validateCert,
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
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
