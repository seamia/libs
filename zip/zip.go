package zip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

const (
	compressedStreamPrefix = 31
)

func Compress(what []byte) []byte {
	var compressed bytes.Buffer
	writer, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if err != nil {
		return nil
	}
	if _, err = writer.Write(what); err != nil {
		return nil
	}
	writer.Close()
	return compressed.Bytes()
}

func Decompress(what []byte) ([]byte, error) {
	if len(what) == 0 {
		// the source is empty - there is nothing here to decompress
		return what, nil
	}

	if what[0] != compressedStreamPrefix {
		// it doesn't seem to be compressed - return the source
		if what[0] != byte('{') && what[0] != byte('[') {
			// fmt.Println("hmmmm.... unexpected prefix of persisted stream ....")
		}
		return what, nil
	}

	gz, err := gzip.NewReader(bytes.NewBuffer(what))
	if err != nil {
		return nil, fmt.Errorf("Read: %v", err)
	}

	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, gz)
	errClose := gz.Close()
	if err != nil {
		return nil, err
	}
	if errClose != nil {
		return nil, errClose
	}

	return decompressed.Bytes(), nil
}
