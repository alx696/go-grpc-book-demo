package main

import (
	"gs/api"
	"gs/filelog"
	toolEnv "gs/tool/env"
	toolSql "gs/tool/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 检查环境变量
	e := toolEnv.Error()
	if e != nil {
		log.Fatalln("环境变量存在问题", e.Error())
	}

	// 初始文件日志
	filelog.Init(toolEnv.SystemId)
	defer func() {
		// 关闭文件日志
		filelog.Over()
	}()

	// 创建postgres连接池
	toolSql.Init()

	// 启动gRPC
	go func() {
		api.Init()
	}()

	// 等待关闭信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	stopSignal := <-signalChan
	filelog.Info("收到关闭信号", stopSignal)

	filelog.Info("服务已经关闭")
}
