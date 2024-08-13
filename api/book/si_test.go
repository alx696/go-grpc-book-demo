package book

import (
	"context"
	"gs/proto/book"
	toolApi "gs/tool/api"
	"log"
	"testing"
	"time"
)

var mc context.Context
var gc book.BookClient

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
	gc = book.NewBookClient(conn)

	m.Run()
}

func TestAdd1(t *testing.T) {
	result, e := gc.Add(mc, &book.Info{
		Code:  "SN1",
		Name:  "软件架构",
		State: "正常",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestAdd2(t *testing.T) {
	result, e := gc.Add(mc, &book.Info{
		Code:  "SN2",
		Name:  "Golang并发",
		State: "正常",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestAdd3(t *testing.T) {
	result, e := gc.Add(mc, &book.Info{
		Code:  "SN3",
		Name:  "架构大师",
		State: "正常",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestChange1(t *testing.T) {
	result, e := gc.Change(mc, &book.Info{
		Code:       "SN1",
		Name:       "软件架构",
		TotalCount: 2,
		State:      "正常",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestChange2(t *testing.T) {
	result, e := gc.Change(mc, &book.Info{
		Code:       "SN2",
		Name:       "Golang进阶",
		TotalCount: 3,
		State:      "正常",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestChange3(t *testing.T) {
	result, e := gc.Change(mc, &book.Info{
		Code:  "SN3",
		Name:  "架构大师",
		State: "删除",
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestSearch(t *testing.T) {
	result, e := gc.Search(mc, &book.SearchRequest{
		PageStart: 1,
		PageCount: 10,
	})

	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}
