package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
	"github.com/goto/salt/printer"
	"github.com/goto/salt/term"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func cmdAction() *cobra.Command {
	var urn, file, output string
	var params structpb.Value
	cmd := &cobra.Command{
		Use:     "action <action-name>",
		Aliases: []string{"action"},
		Short:   "Manage actions",
		Example: heredoc.Doc(`
			$ entropy action start --urn=<resource-urn> --file=<file-path> --out=json
		`),
		Annotations: map[string]string{
			"group:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody entropyv1beta1.ApplyActionRequest
			if file != "" {
				if err := parseFile(file, &params); err != nil {
					return err
				}
				reqBody.Params = &params
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

			fmt.Println(term.Greenf("Action applied successfully"))
			if output == outputJSON || output == outputYAML || output == outputYML {
				formattedString, err := formatOutput(res.GetResource(), output)
				if err != nil {
					return err
				}
				fmt.Println(term.Bluef(formattedString))
			}

			return nil
		}),
	}

	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of the resource")
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to the params file")
	cmd.Flags().StringVarP(&output, "out", "o", "", "output format, `-o json | yaml`")

	return cmd
}
