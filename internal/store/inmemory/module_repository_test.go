package inmemory

import (
	"reflect"
	"testing"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/core/resource/mocks"
	"github.com/odpf/entropy/pkg/errors"
)

func TestModuleRepository_Get(t *testing.T) {
	type fields struct {
		collection map[string]resource.Module
	}
	type args struct {
		id string
	}
	mod := &mocks.Module{}
	mod.EXPECT().ID().Return("mock")

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    resource.Module
		wantErr error
	}{
		{
			name: "Successful",
			fields: fields{
				collection: map[string]resource.Module{
					mod.ID(): mod,
				},
			},
			args: args{
				id: "mock",
			},
			want:    mod,
			wantErr: nil,
		},
		{
			name: "NotFound",
			fields: fields{
				collection: map[string]resource.Module{
					mod.ID(): mod,
				},
			},
			args: args{
				id: "notlog",
			},
			want:    nil,
			wantErr: errors.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &ModuleRepository{
				collection: tt.fields.collection,
			}
			got, err := mr.Get(tt.args.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleRepository_Register(t *testing.T) {
	type fields struct {
		collection map[string]resource.Module
	}
	type args struct {
		module resource.Module
	}
	mod := &mocks.Module{}
	mod.EXPECT().ID().Return("mock")

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "Success",
			fields: fields{
				collection: map[string]resource.Module{},
			},
			args: args{
				module: mod,
			},
			wantErr: nil,
		},
		{
			name: "AlreadyRegistered",
			fields: fields{
				collection: map[string]resource.Module{
					mod.ID(): mod,
				},
			},
			args: args{
				module: mod,
			},
			wantErr: errors.ErrConflict,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &ModuleRepository{
				collection: tt.fields.collection,
			}
			if err := mr.Register(tt.args.module); !errors.Is(err, tt.wantErr) {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
