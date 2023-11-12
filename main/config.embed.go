//go:build feature_embed

package main

import (
	"bytes"
	"crypto/aes"
	"embed"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/xtls/xray-core/common/cmdarg"
	"github.com/xtls/xray-core/common/platform/filesystem"
	"github.com/xtls/xray-core/main/confloader"
)

//go:embed config
var embedConf embed.FS

func init() {
	confloader.EffectiveConfigFileLoader = embedConfigLoader
	configEmbeds = getFileEmbeds()
	filesystem.NewFileReader = embedFileReader
}

func embedConfigLoader(file string) (io.Reader, error) {
	data, err := embedConf.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if strings.Contains(file, "_enc.") {
		raw, err := decodeConfigData(data)
		if err != nil {
			return nil, fmt.Errorf("decode error: %w", err)
		}
		data = stripPad(raw)
	}
	return bytes.NewBuffer(data), nil
}

func embedFileReader(path string) (io.ReadCloser, error) {
	fp, err := embedConf.Open(path)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

func decodeConfigData(data []byte) ([]byte, error) {
	blockSize := 16
	size := len(data)
	if size < blockSize || len(data)%blockSize != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	keys := data[size-blockSize:]
	data = data[:size-blockSize]
	b, err := aes.NewCipher(keys)
	if err != nil {
		return nil, err
	}
	for i := 0; i < size; i += blockSize {
		block := data[i : i+blockSize]
		b.Decrypt(block, block)
	}

	return data, nil
}

func stripPad(data []byte) []byte {
	size := len(data)
	if size > 0 && data[size-1] < 16 {
		strip := int(data[size-1])
		if data[size-strip] != byte(strip) {
			return data
		}
		return data[:size-strip]
	}
	return data
}

func getFileEmbeds() cmdarg.Arg {
	files, err := embedConf.ReadDir("config")
	if err != nil {
		return nil
	}
	var configFiles cmdarg.Arg
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		configFiles = append(configFiles, path.Join("config", file.Name()))
	}
	return configFiles
}
