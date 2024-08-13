package api

import (
	"context"
	"crypto/tls"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// grpc上下文键
type contextKey string

const (
	// 客户端IP
	ContextKeyClientIp contextKey = "client-ip"

	// 用户凭证
	ContextKeyUserToken contextKey = "user-token"

	// 用户ID
	ContextKeyUserId contextKey = "user-id"
)

var (
	// 不要直接使用, 用ProtoToJson方法!
	ProtoJsonMarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	// 不要直接使用, 用ProtoToJsonWithoutDefault方法!
	ProtoJsonMarshalOptionsWithoutDefault = protojson.MarshalOptions{
		EmitUnpopulated: false,
		UseProtoNames:   true,
	}

	// 不要直接使用, 用JsonToProto方法!
	ProtoJsonUnmarshalOptions = protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}
)

// Proto转JSON
func ProtoToJson(m proto.Message) (string, error) {
	jsonBytes, e := ProtoJsonMarshalOptions.Marshal(m)
	if e != nil {
		return "", e
	}
	return string(jsonBytes), nil
}

// Proto转JSON, 不含默认
func ProtoToJsonWithoutDefault(m proto.Message) (string, error) {
	jsonBytes, e := ProtoJsonMarshalOptionsWithoutDefault.Marshal(m)
	if e != nil {
		return "", e
	}
	return string(jsonBytes), nil
}

// JsonToProto JSON转Proto
//
// m 需要指针
func JsonToProto(jsonText string, m proto.Message) error {
	return ProtoJsonUnmarshalOptions.Unmarshal([]byte(jsonText), m)
}

// 获取GRPC Metadata上下文
//
// 注意: 仅供测试用例使用, 作为方法的上下文!
func GetGrpcMetadata(ctx context.Context) context.Context {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		"x-token", os.Getenv("TOKEN"),
	)
	return ctx
}

// 获取GRPC连接
//
// 注意: 仅供测试用例使用, 用完记得关闭连接!
func GetGrpcConn(ctx context.Context) (*grpc.ClientConn, error) {
	var grpcDialOptionArray []grpc.DialOption
	if os.Getenv("GRPC_TLS") == "true" {
		grpcDialOptionArray = append(grpcDialOptionArray, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})))
	} else {
		grpcDialOptionArray = append(grpcDialOptionArray, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	return grpc.NewClient(os.Getenv("GRPC_TARGET"), grpcDialOptionArray...)
}
