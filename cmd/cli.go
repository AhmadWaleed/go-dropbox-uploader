package cmd

import (
	"fmt"
	"os"

	"github.com/AhmadWaleed/dropbox-uploader/pkg/dropbox"
	"github.com/spf13/cobra"
)

var options dropbox.UploadOptions
var token string

var rootCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local file or directory to a remote Dropbox folder.",
	Long: `If the file is bigger than 150Mb the file is uploaded using small chunks (default 50Mb);
		in this case a . (dot) is printed for every chunk successfully uploaded and a * (star) if an 
		error occurs (the upload is retried for a maximum of three times). Only if the file is smaller than 150Mb,
	 	the standard upload API is used, and if the -p option is specified the default curl progress bar is displayed
		 during the upload process. The local file/dir parameter supports wildcards expansion.`,
	Run: func(cmd *cobra.Command, args []string) {
		token := "005ZW5tbmE0AAAAAAAAAAX-fKV88JbsevNEt0cOlncpWvCrSLfmvvhgtDMKByaqF"
		client := dropbox.New(token, options)
		// _ = client
		fmt.Println(options)
		// dropbox.Upload(options)
		if err := client.Dropbox.ChunkedUpload(); err != nil {
			panic(err)
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
