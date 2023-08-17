//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

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

const (
	binDir = "bin"
)

// Build the project and generate binary file.
func Build(_ context.Context) error {
	mg.Deps(Install_deps)

	fmt.Println("️🏗", cyan("building"), green("project"))
	return sh.Run("go", "build", ".")
}

// Install project dependencies.
func Install_deps() error {
	fmt.Println("🚚", cyan("installing"), green("project dependencies"))
	return sh.Run("go", "get", ".")
}

// Generate static files.
func Static_files() error {
	mg.Deps(installTools)

	fmt.Println("🖨️", cyan("generating"), green("static files"))
	return sh.Run("go", "generate", "main.go")
}

// Format project code.
func Format() error {
	mg.Deps(installTools)

	fmt.Println("📐", cyan("formatting"), green("project code"))
	return sh.Run(path.Join(binDir, "gofumpt"), "-w", "-l", ".")
}

// Run tests using `goconvey` tool.
func Test() error {
	mg.Deps(installTools)

	fmt.Println("🏗️", cyan("check"), green("project code"))
	return sh.Run(path.Join(binDir, "goconvey"), "-cover", "-excludedDirs", "bin,build,dist,doc,out,etc,vendor")
}

// Check project code using `golangci-lint` tool.
func Check() error {
	mg.Deps(installTools)

	fmt.Println("🔬", cyan("check"), green("project code"))
	return sh.Run(path.Join(binDir, "golangci-lint"), "run", "./...")
}

// Make linux/amd64 build (for CI and docker).
func Linux_amd64_build() error {
	fmt.Println("🏗️", cyan("building"), green("linux/amd64"))
	return sh.RunWith(
		map[string]string{
			"CGO_ENABLED": "0",
			"GOOS":        "linux",
			"GOARCH":      "amd64",
		},
		"go", "build", ".")
}

// Build docker image for the app.
func Docker() error {
	image := "ccamel/go-graphql-subscription-example"
	fmt.Println("🐳", cyan("dockerize"), green("image"), cyan(image))
	return sh.Run("docker", "build", "-t", image, ".")
}

func installTools() {
	mg.Deps(mg.F(installGoTool, "gofumpt", "mvdan.cc/gofumpt", "v0.5.0"))
	mg.Deps(mg.F(installGoTool, "gothanks", "github.com/psampaz/gothanks", "v0.5.0"))
	mg.Deps(mg.F(installGoTool, "goconvey", "github.com/smartystreets/goconvey", "v1.8.1"))
	mg.Deps(mg.F(installGoTool, "golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint", "v1.54.1"))
}

func installGoTool(name, pkg, version string) error {
	if toolExists(name) {
		return nil
	}

	fmt.Println("🚚", cyan("installing"), green(name), magenta(version))
	binPath, err := filepath.Abs(binDir)
	if err != nil {
		return err
	}
	return sh.RunWith(
		map[string]string{
			"GOBIN": binPath,
		},
		"go", "install", fmt.Sprintf("%s@%s", pkg, version))
}

func toolExists(name string) bool {
	toolPath := filepath.Join(binDir, name)
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		return false
	}
	return true
}
