package env

import (
	"fmt"
	"os"
	"strconv"
)

var (
	// 系统id
	SystemId = os.Getenv("BS_SERVICE_SYSTEM_ID")

	// gRPC端口
	GrpcPort = os.Getenv("BS_SERVICE_GRPC_PORT")

	// postgres连接
	PostgresUri = os.Getenv("POSTGRES_URI")

	// [可选]TLS证书路径
	TlsCrtPath = os.Getenv("TLS_CRT_PATH")
	// [可选]TLS密钥路径
	TlsKeyPath = os.Getenv("TLS_KEY_PATH")
)

// 获取错误
func Error() error {
	if SystemId == "" {
		return fmt.Errorf("没有设置:系统id")
	}

	if GrpcPort == "" {
		return fmt.Errorf("没有设置:gRPC端口")
	}
	_, e := strconv.ParseInt(GrpcPort, 10, 64)
	if e != nil {
		return fmt.Errorf("设置无效:gRPC端口")
	}

	if PostgresUri == "" {
		return fmt.Errorf("没有设置:postgres连接")
	}

	return nil
}
