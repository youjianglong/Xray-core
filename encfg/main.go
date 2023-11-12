package main

import (
	"crypto/aes"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	configDir = "main/config"
	secretKey = "xiaoyao"
	blockSize = 16
	_         = formatKey()
)

func formatKey() struct{} {
	keys := []byte(secretKey)
	if len(keys) >= blockSize {
		keys = keys[:blockSize]
	} else {
		pad := blockSize - len(keys)
		for i := 0; i < pad; i++ {
			keys = append(keys, byte(pad))
		}
	}
	secretKey = string(keys)
	return struct{}{}
}

func encode(data []byte) ([]byte, error) {
	b, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return nil, err
	}
	size := len(data)
	if size%blockSize != 0 {
		pad := blockSize - size%blockSize
		for i := 0; i < pad; i++ {
			data = append(data, byte(pad))
		}
	}
	for i := 0; i < size; i += blockSize {
		block := data[i : i+blockSize]
		b.Encrypt(block, block)
	}
	return data, nil
}

func encodeFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	data, err = encode(data)
	if err != nil {
		return err
	}
	data = append(data, []byte(secretKey)...)
	idx := strings.LastIndexByte(file, '.')
	var filename string
	if idx == -1 {
		filename = file + "_enc."
	} else {
		filename = file[:idx] + "_enc" + file[idx:]
	}
	err = os.WriteFile(filename, data, 0o644)
	if err != nil {
		return err
	}
	_ = os.Remove(file)
	return nil
}

func encodeFiles() error {
	files, err := os.ReadDir(configDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.Contains(file.Name(), "_enc.") {
			continue
		}
		filename := configDir + "/" + file.Name()
		fmt.Println("编码", filename)
		err = encodeFile(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.StringVar(&configDir, "config", configDir, "config dir")
	flag.Parse()
	err := encodeFiles()
	if err != nil {
		panic(err)
	}
	fmt.Println("OK")
}
