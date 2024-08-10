package main

import (
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestServerHandler_OnDescribe(t *testing.T) {
	stream := &gortsplib.ServerStream{}
	var tests = []struct {
		name     string
		response *base.Response
		server   *gortsplib.ServerStream
		err      error
	}{
		{
			name: "OK",
			response: &base.Response{
				StatusCode: base.StatusOK,
			},
			server: stream,
			err:    nil,
		},
		{
			name: "ServerStreamNotExist",
			response: &base.Response{
				StatusCode: base.StatusNotFound,
			},
			server: nil,
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := &serverHandler{}
			sh.stream = tt.server
			r, s, err := sh.OnDescribe(&gortsplib.ServerHandlerOnDescribeCtx{})
			if !gomock.Eq(r).Matches(tt.response) {
				t.Errorf("wrong response %s is not equal to %s", r, tt.response)
			}
			if err != tt.err {
				t.Errorf("wrong error %v is not equal to %v", err, tt.err)
			}
			if s != tt.server {
				t.Errorf("wrong server %v is not equal to %v", s, tt.server)
			}
		})
	}
}

func BenchmarkServerHandler_OnDescribe(b *testing.B) {
	sh := &serverHandler{}
	sh.stream = &gortsplib.ServerStream{}
	for i := 0; i < b.N; i++ {
		_, _, err := sh.OnDescribe(&gortsplib.ServerHandlerOnDescribeCtx{})
		if err != nil {
			b.Error(err.Error())
		}
	}
}

func TestServerHandler_OnSetup(t *testing.T) {
	stream := &gortsplib.ServerStream{}
	var tests = []struct {
		name     string
		response *base.Response
		server   *gortsplib.ServerStream
		err      error
	}{
		{
			name: "OK",
			response: &base.Response{
				StatusCode: base.StatusOK,
			},
			server: stream,
			err:    nil,
		},
		{
			name: "ServerStreamNotExist",
			response: &base.Response{
				StatusCode: base.StatusNotFound,
			},
			server: nil,
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := &serverHandler{}
			sh.stream = tt.server
			r, s, err := sh.OnSetup(&gortsplib.ServerHandlerOnSetupCtx{})
			if !gomock.Eq(r).Matches(tt.response) {
				t.Errorf("wrong response %s is not equal to %s", r, tt.response)
			}
			if err != tt.err {
				t.Errorf("wrong error %v is not equal to %v", err, tt.err)
			}
			if s != tt.server {
				t.Errorf("wrong server %v is not equal to %v", s, tt.server)
			}
		})
	}
}

func BenchmarkServerHandler_OnSetup(b *testing.B) {
	sh := &serverHandler{}
	sh.stream = &gortsplib.ServerStream{}
	for i := 0; i < b.N; i++ {
		_, _, err := sh.OnSetup(&gortsplib.ServerHandlerOnSetupCtx{})
		if err != nil {
			b.Error(err.Error())
		}
	}
}
