package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"FXinnovation/alertmanager-webhook-rocketchat/internal/config"
	"FXinnovation/alertmanager-webhook-rocketchat/internal/server"
	"FXinnovation/alertmanager-webhook-rocketchat/pkg/services/alertprocessor"
	"FXinnovation/alertmanager-webhook-rocketchat/pkg/services/rocketchat"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile    = kingpin.Flag("config.file", "RocketChat configuration file.").Default("config/rocketchat.yml").String()
	listenAddress = kingpin.Flag("listen.address", "The address to listen on for HTTP requests.").Default(":9876").String()
)

func buildContext() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	return fmt.Sprintf("(go=%s)", bi.GoVersion)
}

func buildInfo() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	rev := "unknown"

	for _, info := range bi.Settings {
		if info.Key == "vcs.revision" {
			rev = info.Value
			break
		}
	}

	return fmt.Sprintf("(version=%s, revision=%s)", bi.Main.Version, rev)
}

func RunApp(ctx context.Context) error {
	log.Info("Build context", buildContext())
	log.Info("Starting webhook", buildInfo())

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	go func(cancelFn func(), sigint chan os.Signal) {
		<-sigint
		cancelFn()
	}(cancelFn, sigint)

	cfg, err := config.LoadConfig(*configFile, *listenAddress)
	if err != nil {
		return fmt.Errorf("cannot read config: %w", err)
	}

	rocketChat, err := rocketchat.New(cfg)
	if err != nil {
		return fmt.Errorf("cannot initialize rocketchat client: %w", err)
	}

	ap, err := alertprocessor.New(cfg, rocketChat)
	if err != nil {
		return fmt.Errorf("cannot initialize alert processor: %w", err)
	}

	srv, err := server.New(cfg, ap)
	if err != nil {
		return fmt.Errorf("cannot initialize http server: %w", err)
	}

	shutDownCh := make(chan struct{})
	defer close(shutDownCh)

	if err := srv.Run(ctx, shutDownCh); err != nil {
		return fmt.Errorf("cannot start http server: %w", err)
	}

	<-shutDownCh

	log.Info("shutdown completed")

	return nil
}
