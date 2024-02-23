package docker

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

// Metadata provided by the plugin.
type Metadata struct {
	// SchemaVersion describes the version of this struct. Mandatory, must be "0.1.0"
	SchemaVersion string `json:",omitempty"`
	// Vendor is the name of the plugin vendor. Mandatory
	Vendor string `json:",omitempty"`
	// Version is the optional version of this plugin.
	Version string `json:",omitempty"`
	// ShortDescription should be suitable for a single line help message.
	ShortDescription string `json:",omitempty"`
	// URL is a pointer to the plugin's homepage.
	URL string `json:",omitempty"`
}

func Plugin(vendor, version, desc, url string) *cobra.Command {
	return &cobra.Command{
		Use:   "docker-cli-plugin-metadata",
		Short: "Docker CLI plugin metadata",
		Run: func(cmd *cobra.Command, args []string) {
			metadata := Metadata{
				SchemaVersion:    "0.1.0",
				Vendor:           vendor,
				Version:          version,
				ShortDescription: desc,
				URL:              url,
			}

			json.NewEncoder(cmd.OutOrStdout()).Encode(metadata)
		},
	}
}
