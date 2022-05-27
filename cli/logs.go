package cli

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/odpf/salt/term" //nolint
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func cmdLogs(ctx context.Context) *cobra.Command {
	var urn string
	var filter []string
	filters := make(map[string]string)
	cmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"logs"},
		Short:   "Gets logs",
		Example: heredoc.Doc(`
			$ entropy logs
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			client, cancel, err := createClient(ctx, cmd)
			if err != nil {
				return err
			}
			defer cancel()

			var reqBody entropyv1beta1.GetLogRequest
			for _, f := range filter {
				keyValue := strings.Split(f, ":")
				filters[keyValue[0]] = keyValue[1]
			}

			reqBody.Filter = filters
			reqBody.Urn = urn

			err = reqBody.ValidateAll()
			if err != nil {
				return err
			}

			stream, err := client.GetLog(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			done := make(chan bool)

			go func() {
				for {
					resp, err := stream.Recv()
					if err == io.EOF {
						done <- true //means stream is finished
						return
					}
					if err != nil {
						log.Fatalf("cannot receive %v", err)
					}
					log.Printf(cs.Bluef("%s", resp.GetChunk().GetData()))
				}
			}()

			<-done

			return nil
		},
	}

	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of the resource")
	cmd.Flags().StringArrayVarP(&filter, "filter", "f", nil, "Use filters. Example: --filter=key:value")

	return cmd
}
