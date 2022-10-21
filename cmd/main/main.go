package main

import (
	"context"
	"log"

	"FXinnovation/alertmanager-webhook-rocketchat/internal/app"
	"github.com/prometheus/common/version"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Starts 2 listeners
// - one to give a status on the receiver itself
// - one to actually process the data
func main() {
	kingpin.Version(version.Print("alertmanager-webhook-rocketchat"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	ctx := context.Background()

	err := app.RunApp(ctx)
	if err != nil {
		log.Fatalf("cannot run application: %v", err)
	}
}
