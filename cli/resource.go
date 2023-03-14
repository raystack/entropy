package cli

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/goto/salt/printer"
	"github.com/goto/salt/term" // nolint
	"github.com/spf13/cobra"

	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
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
			$ entropy resource revisions <resource-urn>
		`),
	}

	cmd.AddCommand(
		createResourceCommand(),
		listAllResourcesCommand(),
		viewResourceCommand(),
		editResourceCommand(),
		deleteResourceCommand(),
		getRevisionsCommand(),
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
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody entropyv1beta1.CreateResourceRequest
			if err := parseFile(file, &reqBody); err != nil {
				return err
			} else if err := reqBody.ValidateAll(); err != nil {
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

			fmt.Println("URN: \t", term.Greenf(res.Resource.Urn))
			if output == outputJSON || output == outputYAML || output == outputYML {
				formattedOutput, err := formatOutput(res.GetResource(), output)
				if err != nil {
					return err
				}
				fmt.Println(term.Bluef(formattedOutput))
			}

			return nil
		}),
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
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

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

			if output == outputJSON || output == outputYAML || output == outputYML {
				for _, resource := range res.GetResources() {
					formattedOutput, err := formatOutput(resource, output)
					if err != nil {
						return err
					}
					fmt.Println(term.Bluef(formattedOutput))
				}
			} else {
				var report [][]string
				report = append(report, []string{"URN", "NAME", "KIND", "PROJECT", "STATUS"})
				count := 0
				for _, r := range res.GetResources() {
					report = append(report, []string{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()})
					count++
				}
				printer.Table(os.Stdout, report)
				fmt.Println("\nTotal: ", count)

				fmt.Println(term.Cyanf("To view all the data in JSON/YAML format, use flag `-o json | yaml`"))
			}
			return nil
		}),
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
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

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

			if output == outputJSON || output == outputYAML || output == outputYML {
				formattedOutput, err := formatOutput(res.GetResource(), output)
				if err != nil {
					return err
				}
				fmt.Println(term.Bluef(formattedOutput))
			} else {
				r := res.GetResource()

				printer.Table(os.Stdout, [][]string{
					{"URN", "NAME", "KIND", "PROJECT", "STATUS"},
					{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()},
				})

				fmt.Println(term.Cyanf("\nTo view all the data in JSON/YAML format, use flag `-o json | yaml`"))
			}
			return nil
		}),
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
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var newSpec entropyv1beta1.ResourceSpec
			if err := parseFile(file, &newSpec); err != nil {
				return err
			} else if err := newSpec.ValidateAll(); err != nil {
				return err
			}

			var reqBody entropyv1beta1.UpdateResourceRequest
			reqBody.NewSpec = &newSpec
			reqBody.Urn = args[0]
			if err := reqBody.ValidateAll(); err != nil {
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

			fmt.Println(term.Greenf("Successfully updated"))
			return nil
		}),
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
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

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

			fmt.Println(term.Greenf("Successfully deleted"))
			return nil
		}),
	}
	return cmd
}

func getRevisionsCommand() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "get revisions of a resource",
		Example: heredoc.Doc(`
			$ entropy resource revisions <resource-urn> --out=json
		`),
		Annotations: map[string]string{
			"action:core": "true",
		},
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			spinner := printer.Spin("")
			defer spinner.Stop()

			var reqBody entropyv1beta1.GetResourceRevisionsRequest
			reqBody.Urn = args[0]

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			res, err := client.GetResourceRevisions(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			if output == outputJSON || output == outputYAML || output == outputYML {
				for _, rev := range res.GetRevisions() {
					formattedOutput, err := formatOutput(rev, output)
					if err != nil {
						return err
					}
					fmt.Println(term.Bluef(formattedOutput))
				}
			} else {
				var report [][]string
				report = append(report, []string{"ID", "URN", "CREATED AT"})
				count := 0
				for _, rev := range res.GetRevisions() {
					report = append(report, []string{rev.GetId(), rev.GetUrn(), rev.GetCreatedAt().AsTime().String()})
					count++
				}
				printer.Table(os.Stdout, report)
				fmt.Println("\nTotal: ", count)

				fmt.Println(term.Cyanf("To view all the data in JSON/YAML format, use flag `-o json | yaml`"))
			}
			return nil
		}),
	}

	cmd.Flags().StringVarP(&output, "out", "o", "", "output format, `-o json | yaml`")

	return cmd
}
