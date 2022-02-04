package mongodb

import (
	"errors"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"reflect"
	"testing"
	"time"
)

func TestNewResourceRepository(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	type args struct {
		collection *mongo.Collection
	}
	test := struct {
		name string
		args func(mt *mtest.T) args
		want func(mt *mtest.T) *ResourceRepository
	}{
		name: "test creating new resource repository",
		args: func(mt *mtest.T) args { return args{mt.Coll} },
		want: func(mt *mtest.T) *ResourceRepository { return &ResourceRepository{mt.Coll} },
	}
	mt.Run(test.name, func(mt *mtest.T) {
		if got, want := NewResourceRepository(test.args(mt).collection), test.want(mt); !reflect.DeepEqual(got, want) {
			t.Errorf("NewResourceRepository() = %v, want %v", got, want)
		}
	})
}

func TestResourceRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	type fields struct {
		collection *mongo.Collection
	}
	type args struct {
		resource *domain.Resource
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
				return args{&domain.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Name:      "testname",
					Parent:    "p-testdata-gl",
					Kind:      "log",
					Configs:   map[string]interface{}{},
					Labels:    map[string]string{},
					Status:    domain.ResourceStatusPending,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
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
				return args{&domain.Resource{
					Urn:       "p-testdata-gl-testname-log",
					Name:      "testname",
					Parent:    "p-testdata-gl",
					Kind:      "log",
					Configs:   map[string]interface{}{},
					Labels:    map[string]string{},
					Status:    domain.ResourceStatusPending,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}}
			},
			wantErr: store.ResourceAlreadyExistsError,
		},
	}
	for _, tt := range tests {
		mt.Run(tt.name, func(mt *mtest.T) {
			tt.setup(mt)
			rc := NewResourceRepository(tt.fields(mt).collection)
			if err := rc.Create(tt.args(mt).resource); !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceRepository_GetByURN(t *testing.T) {
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
		want    func(mt *mtest.T) *domain.Resource
		wantErr error
	}{
		{
			name: "test resource get success",
			setup: func(mt *mtest.T) {
				mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.ns", mtest.FirstBatch, bson.D{
					{Key: "urn", Value: "p-testdata-gl-testname-log"},
				}))
			},
			fields:  func(mt *mtest.T) fields { return fields{mt.Coll} },
			args:    func(mt *mtest.T) args { return args{"p-testdata-gl-testname-log"} },
			want:    func(mt *mtest.T) *domain.Resource { return &domain.Resource{Urn: "p-testdata-gl-testname-log"} },
			wantErr: nil,
		},
		{
			name: "test resource not found",
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
			want:    func(mt *mtest.T) *domain.Resource { return nil },
			wantErr: store.ResourceNotFoundError,
		},
	}
	for _, tt := range tests {
		mt.Run(tt.name, func(mt *mtest.T) {
			tt.setup(mt)
			rc := NewResourceRepository(tt.fields(mt).collection)
			got, err := rc.GetByURN(tt.args(mt).urn)
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
