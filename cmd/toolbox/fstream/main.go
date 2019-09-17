package main

import (
	"context"
	"os"

	"github.com/caicloud/nirvana/log"
	"github.com/spf13/cobra"

	"github.com/caicloud/cyclone/pkg/common/signals"
	"github.com/caicloud/cyclone/pkg/util/websocket"
)

var server string
var filePath string

func main() {
	err := newCmd().Execute()
	if err != nil {
		log.Error("Run fstream cmd error:", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// newCmd creates a file stream command
func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fstream",
		Short: "fstream sends file stream to a remote websocket server",
		Long:  "fstream sends file stream to a remote websocket server",
		Run:   run,
	}

	cmd.Flags().StringVarP(&server, "server", "s", "", "url the file stream will send to")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "file path is path of the file")

	return cmd
}

func run(cmd *cobra.Command, args []string) {
	log.Info("Send log stream start")
	file, err := os.Open(filePath)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	ctx, cancel := context.WithCancel(context.Background())
	signals.GracefulShutdown(cancel)

	err = websocket.SendStream(server, file, ctx.Done())
	if err != nil {
		log.Infof("Send file stream %s to %s error:%v", filePath, server, err)
		os.Exit(1)
	}

	log.Info("Send log stream end")
}
