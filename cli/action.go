package cli

import (
	"fmt"

	"github.com/odpf/salt/term" //nolint
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func cmdAction() *cobra.Command {
	var urn, filePath, output string
	var params *structpb.Value
	cmd := &cobra.Command{
		Use:     "action <action-name>",
		Aliases: []string{"action"},
		Short:   "Manage actions",
		Example: heredoc.Doc(`
			$ entropy action start --urn=<resource-urn> --filePath=<file-path> --out=json
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.ApplyActionRequest
			if filePath != "" {
				if err := parseFile(filePath, params); err != nil {
					return err
				}
				reqBody.Params = params
			}

			reqBody.Urn = urn
			reqBody.Action = args[0]

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ApplyAction(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(cs.Greenf("Action applied successfully"))
			if output != "" {
				fmt.Println(cs.Bluef(formatOutput(res.GetResource(), output)))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of the resource")
	cmd.Flags().StringVarP(&filePath, "filePath", "f", "", "path to the params file")
	cmd.Flags().StringVarP(&output, "out", "o", "", "output format, `-o json | yaml`")

	return cmd
}
