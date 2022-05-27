package cli

import (
	"context"
	"fmt"

	"github.com/odpf/salt/term" //nolint
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func cmdAction(ctx context.Context) *cobra.Command {
	var urn, action, filePath string
	var params *structpb.Value
	cmd := &cobra.Command{
		Use:     "action",
		Aliases: []string{"action"},
		Short:   "Manage actions",
		Example: heredoc.Doc(`
			$ entropy action
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
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
			reqBody.Action = action

			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(ctx, cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ApplyAction(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(cs.Greenf("Action applied successfully"))
			fmt.Println(cs.Bluef(prettyPrint(res.GetResource())))

			return nil
		},
	}

	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of the resource")
	cmd.Flags().StringVarP(&action, "action", "a", "", "action to be performed")
	cmd.Flags().StringVarP(&filePath, "filePath", "f", "", "path to the params file")

	return cmd
}
