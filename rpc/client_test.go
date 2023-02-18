package rpc

import (
	"context"
	"net/http"
	"testing"

	"github.com/Jguer/aur"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	newHTTPClient := &http.Client{}

	customRequestEditor := func(ctx context.Context, req *http.Request) error {
		return nil
	}

	type args struct {
		opts []ClientOption
	}
	tests := []struct {
		name             string
		args             args
		wantBaseURL      string
		wanthttpClient   *http.Client
		wantRequestDoers []aur.RequestEditorFn
		wantErr          bool
		wantBatchSize    int
	}{
		{
			name:             "default",
			args:             args{opts: []ClientOption{}},
			wantBaseURL:      "https://aur.archlinux.org/rpc?",
			wanthttpClient:   http.DefaultClient,
			wantErr:          false,
			wantRequestDoers: []aur.RequestEditorFn{},
			wantBatchSize:    150,
		},
		{
			name:             "custom base url",
			args:             args{opts: []ClientOption{WithBaseURL("localhost:8000")}},
			wantBaseURL:      "localhost:8000/rpc?",
			wanthttpClient:   http.DefaultClient,
			wantErr:          false,
			wantRequestDoers: []aur.RequestEditorFn{},
			wantBatchSize:    150,
		},
		{
			name:             "custom base url complete",
			args:             args{opts: []ClientOption{WithBaseURL("localhost:8000/rpc?")}},
			wantBaseURL:      "localhost:8000/rpc?",
			wanthttpClient:   http.DefaultClient,
			wantErr:          false,
			wantRequestDoers: []aur.RequestEditorFn{},
			wantBatchSize:    150,
		},
		{
			name:             "custom http client",
			args:             args{opts: []ClientOption{WithHTTPClient(newHTTPClient)}},
			wantBaseURL:      "https://aur.archlinux.org/rpc?",
			wanthttpClient:   newHTTPClient,
			wantErr:          false,
			wantRequestDoers: []aur.RequestEditorFn{},
			wantBatchSize:    150,
		},
		{
			name:             "custom request editor",
			args:             args{opts: []ClientOption{WithRequestEditorFn(customRequestEditor)}},
			wantBaseURL:      "https://aur.archlinux.org/rpc?",
			wanthttpClient:   newHTTPClient,
			wantErr:          false,
			wantRequestDoers: []aur.RequestEditorFn{customRequestEditor},
			wantBatchSize:    150,
		},
		{
			name:             "want batch size",
			args:             args{opts: []ClientOption{WithBatchSize(300)}},
			wantBaseURL:      "https://aur.archlinux.org/rpc?",
			wanthttpClient:   newHTTPClient,
			wantErr:          false,
			wantRequestDoers: []aur.RequestEditorFn{},
			wantBatchSize:    300,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantBaseURL, got.BaseURL)
			assert.Equal(t, tt.wanthttpClient, got.HTTPClient)
			assert.Equal(t, len(tt.wantRequestDoers), len(got.RequestEditors))
		})
	}
}
