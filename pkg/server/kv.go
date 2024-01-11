package server

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// a simple implementation of the kv store interface
type KV interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	GetAllKeys() []string
	Delete(key string) error
	Close() error
}

type fileKV struct {
	*os.File
	kvmap map[string]string
}

// NewFileKV opens/creates a file and returns a KV. the file is expected to look like this:
// key1=value1
// key2=value2
//
// note that the file is read only once when running this function, and is not dynamically updated by file change
func NewFileKV(path string) (KV, error) {
	// check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// file doesn't exist
		_, err = os.Create(path)
		if err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	// read the file line by line and populate the map
	scanner := bufio.NewScanner(f)
	kvmap := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		// parse the line and populate the map
		// assuming the line format is key=value
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			kvmap[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &fileKV{File: f, kvmap: kvmap}, nil
}

func (k *fileKV) Get(key string) (string, error) {
	value, ok := k.kvmap[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}
	return value, nil
}
func (k *fileKV) GetAllKeys() []string {
	keys := make([]string, 0, len(k.kvmap))
	for key := range k.kvmap {
		keys = append(keys, key)
	}
	return keys
}

func (k *fileKV) Set(key string, value string) error {
	// not implemented in fileKV. the user needs to manually update the file
	return nil
}

func (k *fileKV) Delete(key string) error {
	// not implemented in fileKV. the user needs to manually update the file
	return nil
}

func (k *fileKV) Close() error {
	return k.File.Close()
}
