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
		cloud, _ := cmd.Flags().GetString("cloud")
		prefix, _ := cmd.Flags().GetString("name-prefix")
		region, _ := cmd.Flags().GetString("region")
		ctx := cmd.Context()

		if prefix == "" {
			prefix = "tf-bootstrap"
			log.Printf("No prefix provided. Generated default prefix: %s\n", prefix)
		}
		switch strings.ToLower(cloud) {
		case "aws":
			aws, err := bootstrap.NewAwsBackendService(region)
			if err != nil {
				log.Fatal(err)
			}
			err = aws.Run(ctx, prefix)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("name-prefix", "n", "", "Prefix needed to create a unique name for backend bucket")

	initCmd.Flags().StringP("cloud", "c", "", "Cloud specificied for a bootstrap. aws/gcp/azure")
	if err := initCmd.MarkFlagRequired("cloud"); err != nil {
		log.Println(err)
	}

	initCmd.Flags().StringP("region", "r", "", "Cloud provider region")
	if err := initCmd.MarkFlagRequired("region"); err != nil {
		log.Println(err)
	}
}
