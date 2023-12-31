package commands

import (
	"github.com/kubeTasker/kubeTasker/util/stats"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Load artifacts",
	Run: func(cmd *cobra.Command, args []string) {
		err := loadArtifacts()
		if err != nil {
			log.Fatalf("%+v", err)
		}
	},
}

func loadArtifacts() error {
	wfExecutor := initExecutor()
	defer wfExecutor.HandleError()
	defer stats.LogStats()

	// Download input artifacts
	err := wfExecutor.StageFiles()
	if err != nil {
		wfExecutor.AddError(err)
		return err
	}
	err = wfExecutor.LoadArtifacts()
	if err != nil {
		wfExecutor.AddError(err)
		return err
	}
	return nil
}
