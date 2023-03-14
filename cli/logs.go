package cli

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
	"github.com/goto/salt/term" // nolint

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"
)

func cmdLogs() *cobra.Command {
	var filter []string
	filters := make(map[string]string)
	cmd := &cobra.Command{
		Use:     "logs <resource-urn>",
		Aliases: []string{"logs"},
		Short:   "Gets logs",
		Example: heredoc.Doc(`
			$ entropy logs <resource-urn> --filter="key1=value1" --filter="key2=value2"
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			var reqBody entropyv1beta1.GetLogRequest
			for _, f := range filter {
				keyValue := strings.Split(f, "=")
				filters[keyValue[0]] = keyValue[1]
			}
			reqBody.Filter = filters
			reqBody.Urn = args[0]

			err = reqBody.ValidateAll()
			if err != nil {
				return err
			}

			stream, err := client.GetLog(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			for {
				resp, err := stream.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return fmt.Errorf("failed to read stream: %w", err)
				}

				log.SetFlags(0)
				log.Printf(term.Bluef("%s", resp.GetChunk().GetData())) // nolint
			}

			return nil
		}),
	}

	cmd.Flags().StringArrayVarP(&filter, "filter", "f", nil, "Use filters. Example: --filter=\"key=value\"")
	return cmd
}
