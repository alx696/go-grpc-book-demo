package user

import (
	"context"
	"gs/proto/user"
	toolApi "gs/tool/api"
	"log"
	"testing"
	"time"
)

var mc context.Context
var gc user.UserClient

func TestMain(m *testing.M) {
	// 创建连接
	ctxTimeOut, ctxCancel := context.WithTimeout(context.Background(), time.Second)
	defer ctxCancel()
	conn, e := toolApi.GetGrpcConn(ctxTimeOut)
	if e != nil {
		log.Fatal(e)
	}
	defer conn.Close()

	// 获取Metadata上下文
	mc = toolApi.GetGrpcMetadata(ctxTimeOut)

	// 创建客户端
	gc = user.NewUserClient(conn)

	m.Run()
}

func TestAuth(t *testing.T) {
	result, e := gc.Auth(mc, &user.AuthRequest{
		Username: "测试",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestAdd(t *testing.T) {
	result, e := gc.Add(mc, &user.Info{
		Username: "开发",
		State:    "正常",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestChange(t *testing.T) {
	result, e := gc.Change(mc, &user.Info{
		Username: "开发",
		State:    "删除",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestSearch(t *testing.T) {
	result, e := gc.Search(mc, &user.SearchRequest{
		PageStart: 1,
		PageCount: 10,
	})

	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestSearchWithUsername(t *testing.T) {
	result, e := gc.Search(mc, &user.SearchRequest{
		PageStart: 1,
		PageCount: 10,
		Keyword:   "测",
	})

	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestSearchWithState(t *testing.T) {
	result, e := gc.Search(mc, &user.SearchRequest{
		PageStart: 1,
		PageCount: 10,
		State:     "删除",
	})

	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestExit(t *testing.T) {
	result, e := gc.Exit(mc, &user.Empty{})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}
