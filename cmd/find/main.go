package main

import (
	"fmt"

	"github.com/nomagicln/docker-find/internal"
	"github.com/nomagicln/docker-find/internal/docker"
	"github.com/spf13/cobra"
)

const _1Mib = 1024 * 1024

func FindCommand() *cobra.Command {
	var opts internal.FindOptions
	var noHeader bool

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]

			images, err := internal.FindImages(opts)
			if err != nil {
				return err
			}

			if !noHeader {
				fmt.Printf("%-50s %-20s%-20s\n", "Name", "Size", "Created")
			}

			for _, image := range images {
				fmt.Printf(
					"%-50s %-20s %-20s\n",
					image.Repository+":"+image.Tag,
					fmt.Sprintf("%.2fMib", float64(image.Size)/_1Mib),
					image.Created.Format("2006-01-02 15:04:05"),
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Tag, "tag", "t", "", "The tag of the image")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 10, "The number of images to return")
	cmd.Flags().BoolVarP(&noHeader, "no-header", "H", false, "Hide the header")
	return cmd
}

func main() {
	root := &cobra.Command{Use: "docker-find"}
	root.AddCommand(FindCommand())
	root.AddCommand(docker.Plugin("nomagicln", "0.1.0", "Find images", "https://github.com/nomagicln"))

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
