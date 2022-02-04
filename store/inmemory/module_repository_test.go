package inmemory

import (
	"errors"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/modules/log"
	"github.com/odpf/entropy/store"
	"reflect"
	"testing"
)

func TestModuleRepository_Get(t *testing.T) {
	type fields struct {
		collection map[string]domain.Module
	}
	type args struct {
		id string
	}
	mod := &log.Module{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    domain.Module
		wantErr error
	}{
		{
			name: "test get module from repository",
			fields: fields{
				collection: map[string]domain.Module{
					mod.ID(): mod,
				},
			},
			args: args{
				id: "log",
			},
			want:    mod,
			wantErr: nil,
		},
		{
			name: "test get non-existent module from repository",
			fields: fields{
				collection: map[string]domain.Module{
					mod.ID(): mod,
				},
			},
			args: args{
				id: "notlog",
			},
			want:    nil,
			wantErr: store.ModuleNotFoundError,
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
		collection map[string]domain.Module
	}
	type args struct {
		module domain.Module
	}
	mod := &log.Module{}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "test register module",
			fields: fields{
				collection: map[string]domain.Module{},
			},
			args: args{
				module: mod,
			},
			wantErr: nil,
		},
		{
			name: "test register already added module",
			fields: fields{
				collection: map[string]domain.Module{
					mod.ID(): mod,
				},
			},
			args: args{
				module: mod,
			},
			wantErr: store.ModuleAlreadyExistsError,
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
