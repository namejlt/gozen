package gozen

import (
	"google.golang.org/grpc"
	"testing"
)

func Test_ListenGrpc(t *testing.T) {
	port := ":9999"
	var options []grpc.ServerOption
	s := grpc.NewServer(options...)

	ListenGrpc(port, s)
}
