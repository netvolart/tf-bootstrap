package cmd

import (
	"log"
	"strings"

	"github.com/netvolart/tf-bootstrap/internal/bootstrap"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize backend for a terraform",
	// Long: `You have to provide a Gitlab personal token with --token flag and 
	// GitLab group that is a parrent to your Terraform moduels with --groups flag. 
	// Config will be stored locally`,
	Run: func(cmd *cobra.Command, args []string) {
		prefix, _ := cmd.Flags().GetString("prefix")
		cloud, _ := cmd.Flags().GetString("cloud")

		switch strings.ToLower(cloud) {
		case "aws":
			aws := bootstrap.NewAws(prefix)
			aws.Run()
		}
		
		

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("prefix", "p", "", "Prefix needed to create a unique name for backend bucket")
	if err := initCmd.MarkFlagRequired("prefix"); err != nil {
		log.Println(err)
	}

	initCmd.Flags().StringP("cloud", "c", "", "Cloud specificied for a bootstrap. aws/gcp/azure")
	if err := initCmd.MarkFlagRequired("cloud"); err != nil {
		log.Println(err)
	}
}