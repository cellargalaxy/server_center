package controller

import (
	"context"
	"github.com/cellargalaxy/server_center/model"
	"github.com/cellargalaxy/server_center/service/service"
)

func AddEvent(ctx context.Context, request model.AddEventRequest) (*model.AddEventResponse, error) {
	service.AddEventsAsync(ctx, request.List)
	return &model.AddEventResponse{}, nil
}

func RemoveEvent(ctx context.Context, request model.RemoveEventRequest) (*model.RemoveEventResponse, error) {
	err := service.RemoveEvent(ctx, request.EventInquiry)
	return &model.RemoveEventResponse{}, err
}
