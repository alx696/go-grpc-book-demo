package api

import (
	"context"
	"fmt"
	"gs/api/book"
	"gs/api/user"
	"gs/filelog"
	"net"
	"os"
	"strings"

	toolApi "gs/tool/api"
	toolCache "gs/tool/cache"
	"gs/tool/env"
	toolSql "gs/tool/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// 元数据信息
type metadataInfo struct {
	Ip     string
	Token  string
	ApiKey string
}

var (
	// 公开方法(无需登录)
	publicMethodMap = map[string]int{
		"/user.User/Auth": 1, // 用户:授权
	}
	// 保护方法(需要登录)
	protectMethodMap = map[string]int{
		"/user.User/Exit": 1, // 用户:退出
	}
)

var sdb *pgxpool.Pool

// 获取元数据信息
func getMetadataInfo(ctx context.Context) (*metadataInfo, error) {
	// 获取元数据
	// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md
	// 注意: metadata的键会自动转成全小写字母的!
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("无法读取元数据")
	}

	// 读取客户端IP
	clientIp := ""
	peerInfo, ok := peer.FromContext(ctx)
	if ok {
		peerAddr := peerInfo.Addr.String()

		clientIp = peerAddr[:strings.LastIndex(peerAddr, ":")]

		// 如果来自信任代理, 则从头中提取IP
		if strings.HasPrefix(peerAddr, "172.17.0.1") {
			ipArray := md.Get("x-client-ip")
			if len(ipArray) > 0 {
				clientIp = ipArray[0]
				filelog.Debug("客户端地址", peerAddr, "代理中地址", clientIp)
			}
		}
	}
	mi := metadataInfo{Ip: clientIp}

	// 读取用户凭证
	userTokenArray := md.Get("x-token")
	if len(userTokenArray) > 0 {
		mi.Token = userTokenArray[0]
	}

	// 读取接口密钥
	apiKeyArray := md.Get("x-api-key")
	if len(apiKeyArray) > 0 {
		mi.ApiKey = apiKeyArray[0]
	}

	return &mi, nil
}

// 检查权限
func checkPermission(mi *metadataInfo, fullMethod string) (string, error) {
	// 公开接口
	_, isPublic := publicMethodMap[fullMethod]
	if isPublic {
		filelog.Debug("公开接口", fullMethod)
		return "", nil
	}

	// 使用凭证
	userId := toolCache.GetTokenUser(mi.Token)
	if userId == "" {
		return "", fmt.Errorf("凭证无效")
	}

	// 保护接口
	_, isProtect := protectMethodMap[fullMethod]
	if isProtect {
		filelog.Debug("保护接口", fullMethod)
		return userId, nil
	}

	// TODO 权限接口
	hasApi := true
	if hasApi {
		filelog.Debug("权限接口", fullMethod)
		return userId, nil
	}

	return "", fmt.Errorf("没有权限")
}

// 单式拦截器
// 注意:即使只有返回是流式,这个拦截器还是不处理!
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestJsonText, _ := toolApi.ProtoToJson(req.(proto.Message))

	// 获取元数据
	mi, e := getMetadataInfo(ctx)
	if e != nil {
		filelog.Warn("调用单式接口失败", "接口", info.FullMethod, "请求", requestJsonText, "获取元数据失败", e.Error())
		return nil, status.Errorf(codes.Aborted, e.Error())
	}

	// 检查权限
	userId, e := checkPermission(mi, info.FullMethod)
	if e != nil {
		filelog.Warn("调用单式接口失败", "接口", info.FullMethod, "请求", requestJsonText, "没有权限", e.Error())
		return nil, status.Errorf(codes.PermissionDenied, e.Error())
	}

	// 设置上下文数据
	ctx = context.WithValue(ctx, toolApi.ContextKeyClientIp, mi.Ip)
	ctx = context.WithValue(ctx, toolApi.ContextKeyUserToken, mi.Token)
	ctx = context.WithValue(ctx, toolApi.ContextKeyUserId, userId)

	// 处理
	return handler(ctx, req)
}

// 流式拦截器
// 注意:流式拦截器无法在拦截器中获取请求或返回内容!
func streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	filelog.Info("调用流式接口开始", "接口", info.FullMethod, "暂时不能使用流式接口!")

	return status.Errorf(codes.Unimplemented, "暂时不能使用流式接口!")
}

// Init 初始化
func Init() {
	// 获取服务
	sdb = toolSql.GetDb()

	// 创建服务
	var s *grpc.Server
	var serverOptionArray []grpc.ServerOption
	if env.TlsCrtPath != "" && env.TlsKeyPath != "" {
		filelog.Debug("grpc启用TLS")
		ct, e := credentials.NewServerTLSFromFile(env.TlsCrtPath, env.TlsKeyPath)
		if e != nil {
			filelog.Warn("创建gRPC的TLS证书失败", e)
			os.Exit(1)
		}
		serverOptionArray = append(serverOptionArray, grpc.Creds(ct))
	}
	// https://github.com/grpc/grpc-go/blob/87eb5b7502493f758e76c4d09430c0049a81a557/examples/features/interceptor/server/main.go
	serverOptionArray = append(serverOptionArray, grpc.UnaryInterceptor(unaryInterceptor), grpc.StreamInterceptor(streamInterceptor))
	s = grpc.NewServer(serverOptionArray...)

	// 注册服务
	user.Register(s)
	book.Register(s)

	// 启动服务
	netListen, e := net.Listen("tcp", fmt.Sprint(":", env.GrpcPort))
	if e != nil {
		filelog.Warn("监听gRPC端口失败", e)
		os.Exit(1)
	}
	filelog.Info("启动gRPC服务", "地址", netListen.Addr())
	e = s.Serve(netListen)
	if e != nil {
		filelog.Warn("启动gRPC服务失败", e)
		os.Exit(1)
	}
}
