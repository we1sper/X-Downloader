package main

import (
	"fmt"

	"github.com/we1sper/X-Downloader/pkg/command"
	"github.com/we1sper/X-Downloader/pkg/log"
	"github.com/we1sper/X-Downloader/pkg/util"
	"github.com/we1sper/X-Downloader/x"
	"github.com/we1sper/X-Downloader/x/config"
)

var (
	cfgPath string
	cmd     = command.InitializeCommand()
)

func init() {
	generateConfigTemplateArgument := command.NewMarkArgument("template", "t", "generate a template config file").Action(func() error {
		template := &config.Config{
			Cookie:           "required",
			BearerToken:      "required",
			Retry:            3,
			Downloader:       4,
			Timeout:          60000,
			SaveDir:          "required",
			Overwrite:        false,
			Delta:            true,
			Download:         true,
			Proxy:            "optional",
			LogLevel:         "info",
			LogFile:          "",
			BarrierCandidate: 10,
		}
		if err := util.SaveToJsonFile("./config.json", template); err != nil {
			return fmt.Errorf("failed to generate template config file: %v", err)
		}
		return nil
	})

	configArgument := command.NewValueArgument("config", "c", "specify config file path").Action(func(values []string) error {
		if len(values) > 0 {
			cfgPath = values[0]
		}
		return nil
	})

	screenNameArgument := command.NewValueArgument("user", "u", "specify user screen name").Action(func(values []string) error {
		if len(cfgPath) == 0 {
			return fmt.Errorf("no config file specified")
		}
		if len(values) == 0 {
			return fmt.Errorf("no user specified")
		}
		cfg := config.NewConfig()
		if err := cfg.Load(cfgPath); err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		log.InitializeLog(cfg.LogLevel, cfg.LogFile)
		xClient, err := x.NewXClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize x client: %v", err)
		}
		xClient.Start()
		if _, err = xClient.DownloadMediaTweets(values[0]); err != nil {
			return fmt.Errorf("failed to download tweets: %v", err)
		}
		xClient.Hang()
		return nil
	})

	cmd.Register(generateConfigTemplateArgument).Register(configArgument).Register(screenNameArgument).EnableHelp()
}

func main() {
	cmd.Pipeline()
}
