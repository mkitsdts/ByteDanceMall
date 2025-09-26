package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

func Init() error {
	// 遍历当前文件夹下所有文件
	files, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	path := ""

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".json") {
			path = file.Name()
			slog.Info("Config file found", "path", path)

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(Cfg); err != nil {
				continue
			}
			break
		}
		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			path = file.Name()
			slog.Info("Config file found", "path", path)

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			decoder := yaml.NewDecoder(file)
			if err := decoder.Decode(Cfg); err != nil {
				continue
			}
			break
		}
	}

	return nil
}
