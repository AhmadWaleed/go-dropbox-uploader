package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/AhmadWaleed/dropbox-uploader/pkg/dropbox"
	"github.com/spf13/cobra"
)

var options dropbox.UploadOptions
var token string

var rootCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local file to a remote Dropbox folder.",
	Long:  `If the file is bigger than 150Mb the file is uploaded using small chunks (default 500Mb)`,
	Run: func(cmd *cobra.Command, args []string) {
		file, err := os.Open(options.Source)
		if err != nil {
			log.Fatal(err)
		}

		fileinfo, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		options.File = file
		options.AutoRename = true

		client := dropbox.New(token, options)

		if int(fileinfo.Size()) <= dropbox.UploadFileSizeLimit {
			log.Println("uploading...")
			if err := client.Dropbox.Upload(); err != nil {
				log.Fatal(err)
			}
			return
		}

		log.Println("file size is more then 150 MB, uploading file in chunks.")
		if err := client.Dropbox.ChunkedUpload(); err != nil {
			log.Fatal(err)
		}
	},
}

// Execute root cmd
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&token, "token", "t", "", "dropbox api access token.")
	rootCmd.Flags().StringVarP(&options.Source, "source", "s", "", "local file source path.")
	rootCmd.PersistentFlags().StringVarP(&options.Destination, "destination", "d", "", "dropbox file destination path.")
	rootCmd.PersistentFlags().StringVarP(&options.Mode, "mode", "m", "overwrite", "Selects what to do if the file already exists, default overwite.")
}
