package content

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
)

// X-Content-Type, zip;json, b64;zip;json

// HandlerFromZipJSON decodes a payload containing a pkzip file with a single JSON file inside
func HandlerFromZipJSON(data []byte) ([]byte, error) {
	return unzipFirstFile(data)
}

// HandlerFromZipJSON decodes a base64 payload containing a pkzip file with a single JSON file inside
func HandlerFromBase64ZipJSON(data []byte) ([]byte, error) {
	decoded, err := decodeBase64(data)
	if err != nil {
		return nil, err
	}
	return unzipFirstFile(decoded)
}

func decodeBase64(data []byte) ([]byte, error) {
	var out = make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	_, err := base64.StdEncoding.Decode(out, data)
	if err != nil {
		return out, err
	}
	return out, nil
}

func unzipFirstFile(data []byte) ([]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, file := range z.File {
		data, err := readZipContents(file)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, errors.New("no files found in zip data")
}

func readZipContents(zipFile *zip.File) ([]byte, error) {
	f, err := zipFile.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
