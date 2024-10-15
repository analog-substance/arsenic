package cmd

import (
	"context"
	"fmt"
	"github.com/analog-substance/arsenic/pkg/api"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the arsenic HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		hugoPath, _ := cmd.Flags().GetString("hugo")

		if hugoPath != "" {
			go func() {
				hugoArgs := []string{"server"}
				if len(args) != 0 {
					hugoArgs = append(hugoArgs, args...)
				}

				cmd := exec.CommandContext(context.Background(), "hugo", hugoArgs...)

				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
				cmd.Dir = hugoPath

				err := cmd.Run()
				if err != nil {
					fmt.Println(err)
				}
			}()
		}

		err := api.Serve(port)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 7433, "The port to listen on")
	serveCmd.Flags().StringP("hugo", "H", "", "The path to the hugo directory")
}
