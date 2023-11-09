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
	_ = pflag.Set("logtostderr", "true")
	// We do not want these flags to show up in --help
	_ = pflag.CommandLine.MarkHidden("version")
	_ = pflag.CommandLine.MarkHidden("log-flush-frequency")
	_ = pflag.CommandLine.MarkHidden("alsologtostderr")
	_ = pflag.CommandLine.MarkHidden("log-backtrace-at")
	_ = pflag.CommandLine.MarkHidden("log-dir")
	_ = pflag.CommandLine.MarkHidden("logtostderr")
	_ = pflag.CommandLine.MarkHidden("stderrthreshold")
	_ = pflag.CommandLine.MarkHidden("vmodule")

	cmd := app.NewCommand(version)
	if err := cmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}
