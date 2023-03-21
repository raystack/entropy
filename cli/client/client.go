package client

import (
	"context"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
)

const (
	outFormatFlag   = "format"
	entropyHostFlag = "entropy"
	dialTimeoutFlag = "timeout"

	dialTimeout = 5 * time.Second
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Entropy client with resource management commands",
		Example: heredoc.Doc(`
			$ entropy resource create
			$ entropy resource list
			$ entropy resource view <resource-urn>
			$ entropy resource delete <resource-urn>
			$ entropy resource edit <resource-urn>
			$ entropy resource revisions <resource-urn>
		`),
	}

	cmd.PersistentFlags().StringP(entropyHostFlag, "E", "", "Entropy host to connect to")
	cmd.PersistentFlags().DurationP(dialTimeoutFlag, "T", dialTimeout, "Dial timeout")
	cmd.PersistentFlags().StringP(outFormatFlag, "F", "pretty", "output format (json, yaml, pretty)")

	cmd.AddCommand(
		cmdCreateResource(),
		cmdViewResource(),
		cmdEditResource(),
		cmdStreamLogs(),
		cmdApplyAction(),
		cmdDeleteResource(),
		cmdListRevisions(),
	)

	return cmd
}

func createClient(cmd *cobra.Command) (entropyv1beta1.ResourceServiceClient, func(), error) {
	dialTimeoutVal, _ := cmd.Flags().GetDuration(dialTimeoutFlag)
	entropyAddr, _ := cmd.Flags().GetString(entropyHostFlag)

	dialCtx, dialCancel := context.WithTimeout(cmd.Context(), dialTimeoutVal)
	conn, err := grpc.DialContext(dialCtx, entropyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		dialCancel()
		return nil, nil, err
	}

	cancel := func() {
		dialCancel()
		_ = conn.Close()
	}
	return entropyv1beta1.NewResourceServiceClient(conn), cancel, nil
}
