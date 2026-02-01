package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	_ "unsafe"

	"grpc_server"

	"github.com/matsuridayo/libneko/neko_common"
	"github.com/sagernet/sing-box/constant"
)

func main() {
	fmt.Println("sing-box:", constant.Version, "NekoBox:", neko_common.Version_neko)
	fmt.Println()

	// nekobox_core
	if len(os.Args) > 1 && os.Args[1] == "nekobox" {
		neko_common.RunMode = neko_common.RunMode_NekoBox_Core
		grpc_server.RunCore(setupCore, &server{})
		return
	}

	// Minimal CLI compatibility for scripts (e.g. vpn-run-root.sh).
	if len(os.Args) > 1 && os.Args[1] == "run" {
		os.Exit(runCommand(os.Args[2:]))
	}

	fmt.Fprintf(os.Stderr, "Usage:\n  %s nekobox -port <port> -token <token>\n  %s run -c <config.json>\n", os.Args[0], os.Args[0])
}

func runCommand(args []string) int {
	var configPath string
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&configPath, "c", "config.json", "configuration file path")
	fs.StringVar(&configPath, "config", "config.json", "configuration file path")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read config:", err)
		return 1
	}

	instance, cancel, _, err := createInstance(string(content), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start:", err)
		return 1
	}
	defer func() {
		cancel()
		_ = instance.Close()
	}()

	// Wait for SIGINT/SIGTERM, then shut down.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	<-sigCh
	return 0
}
