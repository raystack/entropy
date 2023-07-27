package serverutils

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const userIDHeader = "user-id"

func GetUserIdentifier(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	xrid := md[userIDHeader]
	if len(xrid) == 0 {
		return "", status.Errorf(codes.InvalidArgument, "missing '%s' header", userIDHeader)
	}

	userID := strings.TrimSpace(xrid[0])
	if userID == "" {
		return "", status.Errorf(codes.InvalidArgument, "empty '%s' header", userIDHeader)
	}

	return userID, nil
}

func ExtractRequestMetadata(_ context.Context, request *http.Request) metadata.MD {
	header := request.Header.Get(userIDHeader)
	md := metadata.Pairs(userIDHeader, header)
	return md
}
