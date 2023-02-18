package rpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Jguer/aur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const errorPayload = `{"version":5,"type":"error","resultcount":0,
"results":[],"error":"Incorrect by field specified."}`

const noMatchPayload = `{"version":5,"type":"error","resultcount":0,
"results":[],"error":""}`

const validPayload = `{
    "version":5,
    "type":"multiinfo",
    "resultcount":1,
    "results":[{
        "ID":229417,
        "Name":"cower",
        "PackageBaseID":44921,
        "PackageBase":"cower",
        "Version":"14-2",
        "Description":"A simple AUR agent with a pretentious name",
        "URL":"http:\/\/github.com\/falconindy\/cower",
        "NumVotes":590,
        "Popularity":24.595536,
        "OutOfDate":null,
        "Maintainer":"falconindy",
		"Submitter":"someone",
        "FirstSubmitted":1293676237,
        "LastModified":1441804093,
        "URLPath":"\/cgit\/aur.git\/snapshot\/cower.tar.gz",
        "Depends":[
            "curl",
            "openssl",
            "pacman",
            "yajl"
        ],
        "MakeDepends":[
            "perl"
        ],
        "License":[
            "MIT"
        ],
        "Keywords":[]
    }]
 }
`

var validPayloadItems = []aur.Pkg{{
	ID: 229417, Name: "cower", PackageBaseID: 44921,
	PackageBase: "cower", Version: "14-2", Description: "A simple AUR agent with a pretentious name",
	URL: "http://github.com/falconindy/cower", NumVotes: 590, Popularity: 24.595536, OutOfDate: 0,
	Maintainer: "falconindy", Submitter: "someone", FirstSubmitted: 1293676237, LastModified: 1441804093,
	URLPath: "/cgit/aur.git/snapshot/cower.tar.gz", Depends: []string{"curl", "openssl", "pacman", "yajl"},
	MakeDepends: []string{"perl"}, CheckDepends: []string(nil), Conflicts: []string(nil),
	Provides: []string(nil), Replaces: []string(nil), OptDepends: []string(nil),
	Groups: []string(nil), License: []string{"MIT"}, Keywords: []string{}, CoMaintainers: []string(nil),
}}

func Test_newAURRPCRequest(t *testing.T) {
	values := url.Values{}
	values.Set("type", "search")
	values.Set("arg", "test-query")
	got, err := newAURRPCRequest(context.Background(), _defaultURL, values)
	assert.NoError(t, err)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg=test-query&type=search&v=5", got.URL.String())
}

func Test_parseRPCResponse(t *testing.T) {
	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name       string
		args       args
		want       []aur.Pkg
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "service unavailable",
			args: args{resp: &http.Response{
				StatusCode: 503,
				Body:       io.NopCloser(bytes.NewBufferString("{}")),
			}},
			want:       []aur.Pkg{},
			wantErr:    true,
			wantErrMsg: "AUR is unavailable at this moment",
		},
		{
			name: "ok empty body",
			args: args{resp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("{}")),
			}},
			want:       nil,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "ok empty body",
			args: args{resp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("{}")),
			}},
			want:       nil,
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "payload error",
			args: args{resp: &http.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewBufferString(errorPayload)),
			}},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "status 400: Incorrect by field specified.",
		},
		{
			name: "valid payload",
			args: args{resp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
			}},
			want:       validPayloadItems,
			wantErr:    false,
			wantErrMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRPCResponse(tt.args.resp)

			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)

				return
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, tt.want, got)
		})
	}
}

type MockedClient struct {
	mock.Mock
}

func (m *MockedClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)

	return args.Get(0).(*http.Response), args.Error(1)
}

func TestClient_Search(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc?",
		HTTPClient:     testClient,
		RequestEditors: []aur.RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
	}, nil)

	got, err := c.Search(context.Background(), "test", aur.Name)

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg=test&by=name&type=search&v=5",
		requestMade.URL.String())
}

func TestClient_Info(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc?",
		HTTPClient:     testClient,
		RequestEditors: []aur.RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
	}, nil)

	got, err := c.Info(context.Background(), []string{"test"})

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}

func TestClient_GetInfo(t *testing.T) {
	testClient := new(MockedClient)

	c, err := NewClient(WithHTTPClient(testClient))
	require.NoError(t, err)

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
	}, nil)

	got, err := c.Get(context.Background(), &aur.Query{
		Needles: []string{"test"},
	})

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}

func TestClient_InfoNoMatch(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc?",
		HTTPClient:     testClient,
		RequestEditors: []aur.RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(noMatchPayload)),
	}, nil)

	got, err := c.Info(context.Background(), []string{"test"})

	assert.NoError(t, err)

	assert.Equal(t, []aur.Pkg{}, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}

func TestClient_InfoError(t *testing.T) {
	testClient := new(MockedClient)

	c := &Client{
		BaseURL:        "https://aur.archlinux.org/rpc?",
		HTTPClient:     testClient,
		RequestEditors: []aur.RequestEditorFn{},
	}

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 503,
		Body:       io.NopCloser(bytes.NewBufferString(errorPayload)),
	}, nil)

	_, err := c.Info(context.Background(), []string{"test"})

	assert.ErrorIs(t, aur.ErrServiceUnavailable, err)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg%5B%5D=test&type=info&v=5",
		requestMade.URL.String())
}

func TestClient_Get(t *testing.T) {
	testClient := new(MockedClient)

	c, err := NewClient(WithHTTPClient(testClient), WithBatchSize(10))
	require.NoError(t, err)

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
	}, nil).Once()

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
	}, nil).Once()

	got, err := c.Get(context.Background(), &aur.Query{
		By:       aur.Name,
		Contains: true,
		Needles:  []string{"test"},
	})

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 2)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg=test&by=name&type=search&v=5",
		requestMade.URL.String())

	requestMadeInfo := testClient.Calls[1].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg%5B%5D=cower&type=info&v=5",
		requestMadeInfo.URL.String())
}

func TestClient_NoBatch(t *testing.T) {
	testClient := new(MockedClient)

	c, err := NewClient(WithHTTPClient(testClient), WithBatchSize(0))
	require.NoError(t, err)

	testClient.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(validPayload)),
	}, nil).Once()

	got, err := c.Get(context.Background(), &aur.Query{
		By:       aur.Name,
		Contains: true,
		Needles:  []string{"test"},
	})

	assert.NoError(t, err)

	assert.Equal(t, validPayloadItems, got)

	testClient.AssertNumberOfCalls(t, "Do", 1)
	testClient.AssertExpectations(t)

	requestMade := testClient.Calls[0].Arguments.Get(0).(*http.Request)
	assert.Equal(t, "https://aur.archlinux.org/rpc?arg=test&by=name&type=search&v=5",
		requestMade.URL.String())
}
