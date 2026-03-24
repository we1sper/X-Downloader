package config

import (
	"fmt"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	cfg := NewConfig()
	if err := cfg.Load("./sample/config.json"); err != nil {
		t.Errorf("load config error: %v", err)
	}
	fmt.Println(cfg)
}
