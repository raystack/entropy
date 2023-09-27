package resources

import (
	"context"

	"go.uber.org/zap"

	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
)

type LogWrapper struct {
	entropyv1beta1.ResourceServiceServer
}

func (lw *LogWrapper) ListResources(ctx context.Context, request *entropyv1beta1.ListResourcesRequest) (*entropyv1beta1.ListResourcesResponse, error) {
	resp, err := lw.ResourceServiceServer.ListResources(ctx, request)
	if err != nil {
		zap.L().Error("ListResources() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (lw *LogWrapper) GetResource(ctx context.Context, request *entropyv1beta1.GetResourceRequest) (*entropyv1beta1.GetResourceResponse, error) {
	resp, err := lw.ResourceServiceServer.GetResource(ctx, request)
	if err != nil {
		zap.L().Error("GetResource() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (lw *LogWrapper) CreateResource(ctx context.Context, request *entropyv1beta1.CreateResourceRequest) (*entropyv1beta1.CreateResourceResponse, error) {
	resp, err := lw.ResourceServiceServer.CreateResource(ctx, request)
	if err != nil {
		zap.L().Error("CreateResource() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (lw *LogWrapper) UpdateResource(ctx context.Context, request *entropyv1beta1.UpdateResourceRequest) (*entropyv1beta1.UpdateResourceResponse, error) {
	resp, err := lw.ResourceServiceServer.UpdateResource(ctx, request)
	if err != nil {
		zap.L().Error("UpdateResource() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (lw *LogWrapper) DeleteResource(ctx context.Context, request *entropyv1beta1.DeleteResourceRequest) (*entropyv1beta1.DeleteResourceResponse, error) {
	resp, err := lw.ResourceServiceServer.DeleteResource(ctx, request)
	if err != nil {
		zap.L().Error("DeleteResource() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (lw *LogWrapper) ApplyAction(ctx context.Context, request *entropyv1beta1.ApplyActionRequest) (*entropyv1beta1.ApplyActionResponse, error) {
	resp, err := lw.ResourceServiceServer.ApplyAction(ctx, request)
	if err != nil {
		zap.L().Error("ApplyAction() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (lw *LogWrapper) GetLog(request *entropyv1beta1.GetLogRequest, server entropyv1beta1.ResourceService_GetLogServer) error {
	err := lw.ResourceServiceServer.GetLog(request, server)
	if err != nil {
		zap.L().Error("GetLog() failed", zap.Error(err))
		return err
	}
	return nil
}

func (lw *LogWrapper) GetResourceRevisions(ctx context.Context, request *entropyv1beta1.GetResourceRevisionsRequest) (*entropyv1beta1.GetResourceRevisionsResponse, error) {
	resp, err := lw.ResourceServiceServer.GetResourceRevisions(ctx, request)
	if err != nil {
		zap.L().Error("GetResourceRevisions() failed", zap.Error(err))
		return nil, err
	}
	return resp, nil
}
