package main

import "github.com/Azure/azure-docker-extension/pkg/vmextension"

type cmdFunc func(vmextension.HandlerEnvironment) error

var (
	cmds = map[string]struct {
		f             cmdFunc // associated function
		name          string  // human readable string
		reportsStatus bool    // determines if running this should log to a .status file
	}{
		"install":   {install, "Install", false},
		"uninstall": {uninstall, "Uninstall", false},
		"enable":    {enable, "Enable", true},
		"update":    {update, "Update", true},
		"disable":   {disable, "Disable", true},
	}
)

func install(vmextension.HandlerEnvironment) error   { return nil }
func uninstall(vmextension.HandlerEnvironment) error { return nil }
func enable(vmextension.HandlerEnvironment) error    { return nil }
func update(vmextension.HandlerEnvironment) error    { return nil }
func disable(vmextension.HandlerEnvironment) error   { return nil }
