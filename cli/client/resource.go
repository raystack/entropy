package client

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goto/salt/printer"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/goto/entropy/pkg/errors"
	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
)

func cmdViewResource() *cobra.Command {
	var kind, project string
	cmd := &cobra.Command{
		Use:     "get [resource-urn]",
		Args:    cobra.MaximumNArgs(1),
		Short:   "List or View existing resource(s)",
		Aliases: []string{"view"},
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			if len(args) == 1 {
				// get resource
				req := entropyv1beta1.GetResourceRequest{
					Urn: args[0],
				}
				spinner := printer.Spin("Getting resource...")
				defer spinner.Stop()
				res, err := client.GetResource(cmd.Context(), &req)
				if err != nil {
					return err
				}
				spinner.Stop()

				r := res.GetResource()
				return Display(cmd, r, func(w io.Writer, v any) error {
					printer.Table(os.Stdout, [][]string{
						{"URN", "NAME", "KIND", "PROJECT", "STATUS"},
						{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()},
					})

					return nil
				})
			}

			// list resource
			req := entropyv1beta1.ListResourcesRequest{
				Kind:    kind,
				Project: project,
			}

			spinner := printer.Spin("Listing resources...")
			defer spinner.Stop()
			res, err := client.ListResources(cmd.Context(), &req)
			if err != nil {
				return err
			}
			spinner.Stop()

			resources := res.GetResources()
			return Display(cmd, resources, func(w io.Writer, _ any) error {
				var report [][]string
				report = append(report, []string{"URN", "NAME", "KIND", "PROJECT", "STATUS"})
				for _, r := range resources {
					report = append(report, []string{r.Urn, r.Name, r.Kind, r.Project, r.State.Status.String()})
				}
				_, _ = fmt.Fprintf(w, "Total: %d\n", len(report)-1)
				printer.Table(os.Stdout, report)
				return nil
			})
		}),
	}

	cmd.Flags().StringVarP(&kind, "kind", "k", "", "kind of resources")
	cmd.Flags().StringVarP(&project, "project", "p", "", "project of resources")

	return cmd
}

func cmdCreateResource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <file>",
		Short: "Create a new resource on Entropy.",
		Args:  cobra.ExactArgs(1),
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			var reqBody entropyv1beta1.Resource
			if err := parseFile(args[0], &reqBody); err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			req := &entropyv1beta1.CreateResourceRequest{
				Resource: &reqBody,
			}

			spinner := printer.Spin("Creating resource...")
			defer spinner.Stop()
			res, err := client.CreateResource(cmd.Context(), req)
			if err != nil {
				return err
			}
			spinner.Stop()

			resource := res.GetResource()
			return Display(cmd, resource, func(w io.Writer, v any) error {
				_, _ = fmt.Fprintf(w, "Resource created with URN '%s'.\n", resource.Urn)
				_, _ = fmt.Fprintln(w, "Use 'entropy resource get <urn>' to view resource.")
				return nil
			})
		}),
	}

	return cmd
}

func cmdEditResource() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "edit <resource-urn>",
		Args:  cobra.ExactArgs(1),
		Short: "Make updates to an existing resource",
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			var newSpec entropyv1beta1.ResourceSpec
			if err := parseFile(file, &newSpec); err != nil {
				return err
			}

			reqBody := entropyv1beta1.UpdateResourceRequest{
				Urn:     args[0],
				NewSpec: &newSpec,
			}
			if err := reqBody.ValidateAll(); err != nil {
				return err
			}

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			spinner := printer.Spin("Updating resource...")
			defer spinner.Stop()
			resp, err := client.UpdateResource(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			resource := resp.GetResource()
			return Display(cmd, resource, func(w io.Writer, _ any) error {
				_, _ = fmt.Fprintln(w, "Update request placed successfully.")
				_, _ = fmt.Fprintln(w, "Use 'entropy resource get <urn>' to view status.")
				return nil
			})
		}),
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "path to the updated spec of resource")
	return cmd
}

func cmdApplyAction() *cobra.Command {
	var urn, file string
	cmd := &cobra.Command{
		Use:     "action <action-name>",
		Args:    cobra.ExactArgs(1),
		Short:   "Apply an action on an existing resource",
		Aliases: []string{"execute"},
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			var params structpb.Value
			if file != "" {
				if err := parseFile(file, &params); err != nil {
					return err
				}
			}

			reqBody := entropyv1beta1.ApplyActionRequest{
				Urn:    urn,
				Action: args[0],
				Params: &params,
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

			spinner := printer.Spin("Applying action...")
			defer spinner.Stop()
			res, err := client.ApplyAction(cmd.Context(), &reqBody)
			if err != nil {
				return err
			}
			spinner.Stop()

			resource := res.GetResource()
			return Display(cmd, resource, func(w io.Writer, v any) error {
				_, _ = fmt.Fprintln(w, "Action request placed successfully.")
				_, _ = fmt.Fprintln(w, "Use 'entropy resource get <urn>' to view status.")
				return nil
			})
		}),
	}

	cmd.Flags().StringVarP(&urn, "urn", "u", "", "urn of the resource")
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to the params file")

	return cmd
}

func cmdDeleteResource() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <resource-urn>",
		Args:    cobra.ExactArgs(1),
		Short:   "Delete an existing resource.",
		Aliases: []string{"rm", "del"},
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			spinner := printer.Spin("Deleting resource...")
			defer spinner.Stop()
			_, err = client.DeleteResource(cmd.Context(), &entropyv1beta1.DeleteResourceRequest{Urn: args[0]})
			if err != nil {
				return err
			}
			spinner.Stop()

			return Display(cmd, nil, func(w io.Writer, v any) error {
				_, _ = fmt.Fprintln(w, "Delete request placed successfully")
				return nil
			})
		}),
	}
	return cmd
}

func cmdListRevisions() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revisions",
		Short:   "List revisions of a resource.",
		Aliases: []string{"revs"},
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			var reqBody entropyv1beta1.GetResourceRevisionsRequest
			reqBody.Urn = args[0]

			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			req := &entropyv1beta1.GetResourceRevisionsRequest{Urn: args[0]}

			spinner := printer.Spin("Retrieving resource revisions...")
			defer spinner.Stop()
			res, err := client.GetResourceRevisions(cmd.Context(), req)
			if err != nil {
				return err
			}
			spinner.Stop()

			revisions := res.GetRevisions()
			return Display(cmd, revisions, func(w io.Writer, v any) error {
				var report [][]string
				report = append(report, []string{"ID", "URN", "CREATED AT"})
				for _, rev := range res.GetRevisions() {
					report = append(report, []string{rev.GetId(), rev.GetUrn(), rev.GetCreatedAt().AsTime().String()})
				}
				printer.Table(os.Stdout, report)
				_, _ = fmt.Fprintf(w, "Total: %d\n", len(report)-1)
				return nil
			})
		}),
	}

	return cmd
}

func cmdStreamLogs() *cobra.Command {
	var filter []string
	cmd := &cobra.Command{
		Use:     "logs <resource-urn>",
		Args:    cobra.ExactArgs(1),
		Short:   "Stream real-time logs for an existing resource.",
		Aliases: []string{"logs"},
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			client, cancel, err := createClient(cmd)
			if err != nil {
				return err
			}
			defer cancel()

			filters := map[string]string{}
			for _, f := range filter {
				keyValue := strings.Split(f, ":")
				filters[keyValue[0]] = keyValue[1]
			}

			reqBody := &entropyv1beta1.GetLogRequest{
				Urn:    args[0],
				Filter: filters,
			}

			if err := reqBody.ValidateAll(); err != nil {
				return err
			}

			spinner := printer.Spin("Preparing to stream logs...")
			defer spinner.Stop()
			stream, err := client.GetLog(cmd.Context(), reqBody)
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

				chunk := resp.GetChunk()
				_ = Display(cmd, chunk, func(w io.Writer, v any) error {
					_, _ = fmt.Fprintln(w, string(chunk.GetData()))
					return nil
				})
			}

			return nil
		}),
	}

	cmd.Flags().StringSliceVarP(&filter, "filter", "f", nil, "Filter. (e.g., --filter=\"key:value\")")
	return cmd
}
