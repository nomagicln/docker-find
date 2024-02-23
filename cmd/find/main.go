package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/nomagicln/docker-find/internal"
	"github.com/nomagicln/docker-find/internal/docker"
	"github.com/spf13/cobra"
)

const _1Mib = 1024 * 1024

func FindCommand() *cobra.Command {
	var (
		opts       internal.FindOptions
		noHeader   bool
		showDetail bool
		onlyDetail bool
	)

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]

			if showDetail {
				detail, err := internal.GetDetail(opts.Name)
				if err != nil {
					return err
				}

				s, err := glamour.Render(detail.FullDescription, "dark")
				if err != nil {
					return err
				}

				fmt.Println("Description: ")
				fmt.Println(s)

				if onlyDetail {
					return nil
				}

				fmt.Println("\nImages: ")
			}

			var (
				ctlCtx, cancel = context.WithCancel(cmd.Context())
				ctlCh          = make(chan string)
				images         []internal.Image
				next           internal.FindFunc = internal.FindImagesFunc(ctlCtx, opts)
				err            error
			)

			if !noHeader {
				fmt.Printf("%-50s %-20s%-20s\n", "Name", "Size", "Created")
			}

			pr, pw := io.Pipe()

			go func() {
				<-ctlCtx.Done()
				pw.Close()
				os.Exit(0)
			}()

			go func() {
				defer cancel()

				buf := bufio.NewReader(os.Stdin)
				for {
					b, err := buf.ReadByte()
					if err != nil {
						if err == io.EOF {
							break
						}
						fmt.Fprintf(os.Stderr, "failed to read from stdin: %v\n", err)
						return
					}

					var s string
					switch b {
					case 'q':
						s = "q"
					case '\n', 'n', ' ', 'j', 'k', '\r':
						s = "n"
					default:
						continue
					}
					ctlCh <- s
				}
			}()

			go func() {
				defer cancel()

				for next != nil {
					select {
					case s := <-ctlCh:
						switch s {
						case "q":
							return

						case "n":
							images, next, err = next()
							if err != nil {
								fmt.Fprintf(os.Stderr, "failed to fetch images: %v\n", err)
								return
							}
						}
					case <-ctlCtx.Done():
						return
					}

					for _, image := range images {
						fmt.Fprintf(
							pw,
							"%-50s %-20s %-20s\n",
							image.Repository+":"+image.Tag,
							fmt.Sprintf("%.2fMib", float64(image.Size)/_1Mib),
							image.Created.Format("2006-01-02 15:04:05"),
						)
					}
				}
			}()

			ctlCh <- "n"

			pagerCmd := os.Getenv("PAGER")
			if pagerCmd == "" {
				pagerCmd = "less"
			}

			pa := strings.Split(pagerCmd, " ")
			c := exec.CommandContext(ctlCtx, pa[0], pa[1:]...) // nolint:gosec
			c.Stdin = pr
			c.Stdout = os.Stdout
			if err := c.Run(); err != nil && ctlCtx.Err() != context.Canceled {
				fmt.Fprintf(os.Stderr, "failed to run pager: %v\n", err)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&opts.Page, "page", "p", 1, "The page number")
	cmd.Flags().IntVarP(&opts.PageSize, "page-size", "P", 25, "The number of images to return per page")
	cmd.Flags().BoolVarP(&noHeader, "no-header", "H", false, "Hide the header")
	cmd.Flags().BoolVarP(&showDetail, "detail", "d", false, "Show detail of the image")
	cmd.Flags().BoolVarP(&onlyDetail, "only-detail", "D", false, "Only show detail of the image")

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
