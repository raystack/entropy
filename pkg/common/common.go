package common

import (
	"context"

	commonv1 "github.com/goto/entropy/proto/gotocompany/common/v1"
)

type CommonService struct {
	commonv1.UnimplementedCommonServiceServer
	version *commonv1.Version
}

func New(version *commonv1.Version) *CommonService {
	return &CommonService{
		version: version,
	}
}

func (c *CommonService) GetVersion(context.Context, *commonv1.GetVersionRequest) (*commonv1.GetVersionResponse, error) {
	return &commonv1.GetVersionResponse{
		Server: c.version,
	}, nil
}
