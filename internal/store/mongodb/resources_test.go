package mongodb

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func TestResourceStore_Create(t *testing.T) {
	t.Parallel()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	type fields struct {
		collection *mongo.Collection
	}
	type args struct {
		resource resource.Resource
	}
	tests := []struct {
		name    string
		setup   func(mt *mtest.T)
		fields  func(mt *mtest.T) fields
		args    func(mt *mtest.T) args
		wantErr error
	}{
		{
			name:   "test create success",
			setup:  func(mt *mtest.T) { mt.AddMockResponses(mtest.CreateSuccessResponse()) },
			fields: func(mt *mtest.T) fields { return fields{mt.Coll} },
			args: func(mt *mtest.T) args {
				return args{resource.Resource{
					URN:       "p-testdata-gl-testname-log",
					Kind:      "log",
					Name:      "testname",
					Project:   "p-testdata-gl",
					Labels:    map[string]string{},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Spec: resource.Spec{
						Configs: []byte("{}"),
					},
					State: resource.State{
						Status: resource.StatusPending,
					},
				}}
			},
			wantErr: nil,
		},
		{
			name: "test create duplicate",
			setup: func(mt *mtest.T) {
				mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
					Index:   1,
					Code:    11000,
					Message: "duplicate key error",
				}))
			},
			fields: func(mt *mtest.T) fields {
				return fields{mt.Coll}
			},
			args: func(mt *mtest.T) args {
				return args{resource.Resource{
					URN:       "p-testdata-gl-testname-log",
					Name:      "testname",
					Kind:      "log",
					Project:   "p-testdata-gl",
					Labels:    map[string]string{},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Spec: resource.Spec{
						Configs: []byte("{}"),
					},
					State: resource.State{
						Status: resource.StatusPending,
					},
				}}
			},
			wantErr: errors.ErrConflict,
		},
	}
	for _, tt := range tests {
		mt.Run(tt.name, func(mt *mtest.T) {
			tt.setup(mt)
			rc := NewResourceStore(mt.DB)
			if err := rc.Create(context.Background(), tt.args(mt).resource); !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceStore_GetByURN(t *testing.T) {
	t.Parallel()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	type fields struct {
		collection *mongo.Collection
	}
	type args struct {
		urn string
	}
	tests := []struct {
		name    string
		setup   func(mt *mtest.T)
		fields  func(mt *mtest.T) fields
		args    func(mt *mtest.T) args
		want    func(mt *mtest.T) *resource.Resource
		wantErr error
	}{
		{
			name: "Success",
			setup: func(mt *mtest.T) {
				mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.ns", mtest.FirstBatch, bson.D{
					{Key: "urn", Value: "p-testdata-gl-testname-log"},
				}))
			},
			fields:  func(mt *mtest.T) fields { return fields{mt.Coll} },
			args:    func(mt *mtest.T) args { return args{"p-testdata-gl-testname-log"} },
			want:    func(mt *mtest.T) *resource.Resource { return &resource.Resource{URN: "p-testdata-gl-testname-log"} },
			wantErr: nil,
		},
		{
			name: "NotFound",
			setup: func(mt *mtest.T) {
				mt.AddMockResponses(bson.D{
					{Key: "ok", Value: 1},
					{Key: "cursor", Value: bson.D{
						{Key: "id", Value: int64(0)},
						{Key: "ns", Value: "test.ns"},
						{Key: string(mtest.FirstBatch), Value: bson.A{}},
					}},
				})
			},
			fields:  func(mt *mtest.T) fields { return fields{mt.Coll} },
			args:    func(mt *mtest.T) args { return args{"p-testdata-gl-unknown-log"} },
			want:    func(mt *mtest.T) *resource.Resource { return nil },
			wantErr: errors.ErrNotFound,
		},
	}
	for _, tt := range tests {
		mt.Run(tt.name, func(mt *mtest.T) {
			tt.setup(mt)
			rc := NewResourceStore(mt.DB)
			got, err := rc.GetByURN(context.Background(), tt.args(mt).urn)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetByURN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if want := tt.want(mt); !reflect.DeepEqual(got, want) {
				t.Errorf("GetByURN() got = %v, want %v", got, want)
			}
		})
	}
}
