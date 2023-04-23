package main

import (
	"flag"
	"os"

	"github.com/crazytaxii/volume-qos-controller/cmd/qos-controller/app"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

var version string

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Set("logtostderr", "true")
	// We do not want these flags to show up in --help
	pflag.CommandLine.MarkHidden("version")
	pflag.CommandLine.MarkHidden("log-flush-frequency")
	pflag.CommandLine.MarkHidden("alsologtostderr")
	pflag.CommandLine.MarkHidden("log-backtrace-at")
	pflag.CommandLine.MarkHidden("log-dir")
	pflag.CommandLine.MarkHidden("logtostderr")
	pflag.CommandLine.MarkHidden("stderrthreshold")
	pflag.CommandLine.MarkHidden("vmodule")

	cmd := app.NewCommand(version)
	if err := cmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}
