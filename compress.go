package goutils

import (
	"bytes"
	"compress/gzip"
)

func CompressWithGzip(data []byte) (compressedData []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err = gz.Write(data)
	if err != nil {
		_ = gz.Close()

		return
	}

	// 显式 Close，确保 gzip footer 和所有缓冲数据都写入 bytes.Buffer
	err = gz.Close()
	if err != nil {
		return
	}

	// Close 之后再读取，才是完整的压缩数据
	compressedData = b.Bytes()

	return
}
