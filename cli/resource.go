package cli

import (
	"fmt"
	"os"

	"github.com/odpf/salt/term" //nolint
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	"github.com/MakeNowJust/heredoc"
	"github.com/odpf/salt/printer"
	"github.com/spf13/cobra"
)

func cmdResource() *cobra.Command {
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
			$ entropy resource view <resource-urn>
			$ entropy resource delete <resource-urn>
			$ entropy resource edit <resource-urn>
		`),
	}

	cmd.AddCommand(
		createResourceCommand(),
		listAllResourcesCommand(),
		viewResourceCommand(),
		editResourceCommand(),
		deleteResourceCommand(),
	)

	return cmd
}

func createResourceCommand() *cobra.Command {
	var file, output string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a resource",
		Example: heredoc.Doc(`
			$ entropy resource create --file=<file-path> --out=json
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.CreateResourceRequest
			if err := parseFile(file, &reqBody); err != nil {
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

			res, err := client.CreateResource(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println("URN: \t", cs.Greenf(res.Resource.Urn))
			if output != "" {
				fmt.Println(cs.Bluef(formatOutput(res.GetResource(), output)))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "path to body of resource")
	cmd.Flags().StringVarP(&output, "out", "o", "", "output format, `-o json | yaml`")

	return cmd
}

func listAllResourcesCommand() *cobra.Command {
	var output, kind, project string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all resources",
		Example: heredoc.Doc(`
			$ entropy resource list --kind=<resource-kind> --project=<project-name> --out=json
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

			res, err := client.ListResources(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			if output != "" {
				for _, resource := range res.GetResources() {
					fmt.Println(cs.Bluef(formatOutput(resource, output)))
				}
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

				fmt.Println(cs.Cyanf("To view all the data in JSON/YAML format, use flag `-o json | yaml`"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "out", "o", "", "output format, `-o json | yaml`")
	cmd.Flags().StringVarP(&kind, "kind", "k", "", "kind of resources")
	cmd.Flags().StringVarP(&project, "project", "p", "", "project of resources")

	return cmd
}

func viewResourceCommand() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "view <resource-urn>",
		Short: "view a resource",
		Example: heredoc.Doc(`
			$ entropy resource view <resource-urn> --out=json
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.GetResourceRequest
			reqBody.Urn = args[0]

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()
			res, err := client.GetResource(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			if output != "" {
				fmt.Println(cs.Bluef(formatOutput(res.GetResource(), output)))
			} else {
				report := [][]string{}
				report = append(report, []string{"URN", "NAME", "KIND", "PROJECT", "STATUS"})
				r := res.GetResource()
				report = append(report, []string{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()})

				printer.Table(os.Stdout, report)

				fmt.Println(cs.Cyanf("\nTo view all the data in JSON/YAML format, use flag `-o json | yaml`"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "out", "o", "", "output format, `-o json | yaml`")

	return cmd
}

func editResourceCommand() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "edit <resource-urn>",
		Short: "edit a resource",
		Example: heredoc.Doc(`
			$ entropy resource edit <resource-urn> --file=<file-path>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var newSpec entropyv1beta1.ResourceSpec
			if err := parseFile(file, &newSpec); err != nil {
				return err
			}
			err := newSpec.ValidateAll()
			if err != nil {
				return err
			}

			var reqBody entropyv1beta1.UpdateResourceRequest
			reqBody.NewSpec = &newSpec
			reqBody.Urn = args[0]
			err = reqBody.ValidateAll()
			if err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.UpdateResource(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(cs.Greenf("Successfully updated"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "path to the updated spec of resource")

	return cmd
}

func deleteResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <resource-urn>",
		Short: "delete a resource",
		Example: heredoc.Doc(`
			$ entropy resource delete <resource-urn>
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()
			cs := term.NewColorScheme()

			var reqBody entropyv1beta1.DeleteResourceRequest
			reqBody.Urn = args[0]

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			_, err = client.DeleteResource(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			fmt.Println(cs.Greenf("Successfully deleted"))
			return nil
		},
	}
	return cmd
}
