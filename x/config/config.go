package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Cookie           string
	BearerToken      string
	Retry            uint64
	Downloader       uint64
	Timeout          uint64
	SaveDir          string
	Overwrite        bool
	Delta            bool
	Download         bool
	Proxy            string
	LogLevel         string
	LogFile          string
	BarrierCandidate uint64
}

func NewConfig() *Config {
	return &Config{
		Retry:            5,
		Downloader:       4,
		Timeout:          60000,
		Delta:            true,
		Download:         true,
		LogLevel:         "info",
		BarrierCandidate: 10,
	}
}

func (cfg *Config) Load(path string) error {
	bytes, readErr := os.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	unmarshalErr := json.Unmarshal(bytes, cfg)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	return nil
}

func (cfg *Config) String() string {
	str := "Config:\n"
	str += fmt.Sprintf("Cookie=%v\n", cfg.Cookie)
	str += fmt.Sprintf("BearerToken=%v\n", cfg.BearerToken)
	str += fmt.Sprintf("Retry=%v\n", cfg.Retry)
	str += fmt.Sprintf("Downloader=%v\n", cfg.Downloader)
	str += fmt.Sprintf("Timeout=%v\n", cfg.Timeout)
	str += fmt.Sprintf("SaveDir=%v\n", cfg.SaveDir)
	str += fmt.Sprintf("Overwrite=%v\n", cfg.Overwrite)
	str += fmt.Sprintf("Delta=%v\n", cfg.Delta)
	str += fmt.Sprintf("Download=%v\n", cfg.Download)
	str += fmt.Sprintf("Proxy=%v\n", cfg.Proxy)
	str += fmt.Sprintf("LogLevel=%v\n", cfg.LogLevel)
	str += fmt.Sprintf("LogFile=%v\n", cfg.LogFile)
	str += fmt.Sprintf("BarrierCandidate=%v\n", cfg.BarrierCandidate)
	return str
}
