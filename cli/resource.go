package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/odpf/salt/term" //nolint
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

const outputJSON = "json"

func cmdResource(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"resources"},
		Short:   "Manage resources",
		Annotations: map[string]string{
			"group:core": "true",
		},
		Example: heredoc.Doc(`
			$ entropy resource create
			$ entropy resource list
			$ entropy resource view
			$ entropy resource delete
			$ entropy resource edit
		`),
	}

	cmd.AddCommand(
		createResourceCmd(ctx),
		listAllResourcesCommand(ctx),
		viewResourceCommand(ctx),
		editResourceCommand(ctx),
		deleteResourceCommand(ctx),
	)

	return cmd
}

func createResourceCmd(ctx context.Context) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a resource",
		Example: heredoc.Doc(`
			$ entropy resource create
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.CreateResourceRequest
			if err := parseFile(filePath, &reqBody); err != nil {
				return err
			}
			err := reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.CreateResource(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("URN: \t", cs.Greenf(res.Resource.Urn))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "filePath", "f", "", "path to body of resource")

	return cmd
}

func listAllResourcesCommand(ctx context.Context) *cobra.Command {
	var output, kind, project string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all resources",
		Example: heredoc.Doc(`
			$ entropy resource list
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.ListResourcesRequest
			reqBody.Kind = kind
			reqBody.Project = project

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.ListResources(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			if output == outputJSON {
				fmt.Println(cs.Bluef(prettyPrint(res.GetResources())))
			} else {
				report := [][]string{}
				report = append(report, []string{"URN", "NAME", "KIND", "PROJECT", "STATUS"})
				count := 0
				for _, r := range res.GetResources() {
					report = append(report, []string{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()})
					count++
				}
				printer.Table(os.Stdout, report)
				fmt.Println("\nTotal: ", count)

				fmt.Println(cs.Cyanf("To view all the data in JSON format, use flag `-o json`"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "out", "o", "table", "output format, for json `-o json`")
	cmd.Flags().StringVarP(&kind, "kind", "k", "", "kind of resources")
	cmd.Flags().StringVarP(&project, "project", "p", "", "project of resources")

	return cmd
}

func viewResourceCommand(ctx context.Context) *cobra.Command {
	var output, urn string
	cmd := &cobra.Command{
		Use:   "view",
		Short: "view a resource",
		Example: heredoc.Doc(`
			$ entropy resource view 
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.GetResourceRequest
			reqBody.Urn = urn

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()
			res, err := client.GetResource(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			if output == outputJSON {
				fmt.Println(cs.Bluef(prettyPrint(res.GetResource())))
			} else {
				report := [][]string{}
				report = append(report, []string{"URN", "NAME", "KIND", "PROJECT", "STATUS"})
				r := res.GetResource()
				report = append(report, []string{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()})

				printer.Table(os.Stdout, report)

				fmt.Println(cs.Cyanf("\nTo view all the data in JSON format, use flag `-o json`"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "out", "o", "table", "output format, for json `-o json`")
	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of resource")

	return cmd
}

func editResourceCommand(ctx context.Context) *cobra.Command {
	var filePath, urn string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edit a resource",
		Example: heredoc.Doc(`
			$ entropy resource edit
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var newSpec entropyv1beta1.ResourceSpec
			if err := parseFile(filePath, &newSpec); err != nil {
				return err
			}
			err := newSpec.ValidateAll()
			if err != nil {
				return err
			}

			var reqBody entropyv1beta1.UpdateResourceRequest
			reqBody.NewSpec = &newSpec
			reqBody.Urn = urn
			err = reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.UpdateResource(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(cs.Greenf("Successfully updated"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "filePath", "f", "", "path to the updated spec of resource")
	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of resource")

	return cmd
}

func deleteResourceCommand(ctx context.Context) *cobra.Command {
	var urn string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete a resource",
		Example: heredoc.Doc(`
			$ entropy resource delete 
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.DeleteResourceRequest
			reqBody.Urn = urn

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.DeleteResource(ctx, &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(cs.Greenf("Successfully deleted"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of resource")

	return cmd
}
