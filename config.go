package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
)

// Config Service configuration
type Config struct {
	Path     string
	Pattern  string
	Password string
}

func getConfigPath() (string, error) {
	fullexecpath, err := osext.Executable()
	if err != nil {
		return "", err
	}

	dir, execname := filepath.Split(fullexecpath)
	ext := filepath.Ext(execname)
	name := execname[:len(execname)-len(ext)]

	return filepath.Join(dir, name+".json"), nil
}

func loadConfig(path string) (*Config, error) {
	dir, _ := filepath.Split(path)

	config := &Config{
		Path:     dir,
		Pattern:  "*.pdf",
		Password: "1234",
	}

	if _, err := os.Stat(path); err == nil {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		r := json.NewDecoder(f)
		err = r.Decode(&config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func getConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}
