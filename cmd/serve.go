package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/analog-substance/arsenic/api"
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
				cmd := exec.CommandContext(context.Background(), "hugo", "server")

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
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 7433, "The port to listen on")
	serveCmd.Flags().StringP("hugo", "H", "", "The path to the hugo directory")
}
