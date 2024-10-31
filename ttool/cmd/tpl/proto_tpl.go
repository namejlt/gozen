package tpl

var (
	ProtoDirName   = "proto"
	ProtoFilesName = []string{
		"readme.md",
		"test.proto",
	}
	ProtoFilesContent = []string{
		protoReadmeMd,
		protoTestProto,
	}
)

var (
	protoReadmeMd = `获取grpc代码

protoc  --go_out=. test.proto

protoc  --go-grpc_out=. test.proto
`
	protoTestProto = `syntax = "proto3";

package test;

option go_package = "grpc/test";

service Test {
    rpc Show (ShowRequest) returns (ShowResponse) {
    }
}

//基础响应结构
message BaseResponse {
    int32 code = 1;
    string message = 2;
}

//Show
message ShowRequest {
    string content = 1;
}

message ShowResponse {
    BaseResponse info = 1;
    ShowResponseData data = 2;
}

message ShowResponseData {
    string content = 1;
}
`
)
