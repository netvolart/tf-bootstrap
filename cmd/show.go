package cmd

import (
	"log"
	"strings"

	"github.com/netvolart/tf-bootstrap/internal/bootstrap"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Initialize backend for a terraform",
	// Long: `You have to provide a Gitlab personal token with --token flag and
	// GitLab group that is a parrent to your Terraform moduels with --groups flag.
	// Config will be stored locally`,
	Run: func(cmd *cobra.Command, args []string) {
		cloud, _ := cmd.Flags().GetString("cloud")
		region, _ := cmd.Flags().GetString("region")
		ctx := cmd.Context()

		switch strings.ToLower(cloud) {
		case "aws":
			aws, err := bootstrap.NewAwsBackendService(region)
			if err != nil {
				log.Fatal(err)
			}
			err = aws.Show(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().StringP("region", "r", "", "Cloud provider region")
	if err := initCmd.MarkFlagRequired("region"); err != nil {
		log.Println(err)
	}

	showCmd.Flags().StringP("cloud", "c", "", "Cloud specificied for a bootstrap. aws/gcp/azure")
	if err := initCmd.MarkFlagRequired("cloud"); err != nil {
		log.Println(err)
	}
}
