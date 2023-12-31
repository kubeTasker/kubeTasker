// Package cmd provides functionally common to various kubeTasker CLIs

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/user"
	"strings"

	kubeTasker "github.com/kubeTasker/kubeTasker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewVersionCmd returns a new `version` command to be used as a sub-command to root
func NewVersionCmd(cliName string) *cobra.Command {
	var short bool
	versionCmd := cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			version := kubeTasker.GetVersion()
			fmt.Printf("%s: %s\n", cliName, version)
			if short {
				return
			}
			fmt.Printf("  BuildDate: %s\n", version.BuildDate)
			fmt.Printf("  GitCommit: %s\n", version.GitCommit)
			fmt.Printf("  GitTreeState: %s\n", version.GitTreeState)
			if version.GitTag != "" {
				fmt.Printf("  GitTag: %s\n", version.GitTag)
			}
			fmt.Printf("  GoVersion: %s\n", version.GoVersion)
			fmt.Printf("  Compiler: %s\n", version.Compiler)
			fmt.Printf("  Platform: %s\n", version.Platform)
		},
	}
	versionCmd.Flags().BoolVar(&short, "short", false, "print just the version number")
	return &versionCmd
}

// MustIsDir returns whether or not the given filePath is a directory. Exits if path does not exist
func MustIsDir(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return fileInfo.IsDir()
}

// MustHomeDir returns the home directory of the user
func MustHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

// IsURL returns whether or not a string is a URL
func IsURL(u string) bool {
	var parsedURL *url.URL
	var err error

	parsedURL, err = url.ParseRequestURI(u)
	if err == nil {
		if parsedURL != nil && parsedURL.Host != "" {
			return true
		}
	}
	return false
}

// SetLogLevel sets the logrus logging level
func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalf("Unknown level: %s", level)
	}
}
