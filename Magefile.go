//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	magenta = color.New(color.FgMagenta).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
)

func installTools() {
	mg.Deps(mg.F(installGoTool, "gofumpt", "mvdan.cc/gofumpt", "v0.5.0"))
	mg.Deps(mg.F(installGoTool, "gothanks", "psampaz/gothanks", "latest"))
	mg.Deps(mg.F(installGoTool, "goconvey", "smartystreets/goconvey", "latest"))
	mg.Deps(mg.F(installGoTool, "golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint", "v1.54.1"))
}

func installGoTool(name, p, version string) error {
	if _, err := exec.LookPath(name); err == nil {
		return nil
	}

	fmt.Println("🚚", cyan("installing"), green(name), magenta(version))
	return sh.Run("go", "install", fmt.Sprintf("%s@%s", p, version))
}
