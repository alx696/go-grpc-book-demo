package borrow

import (
	"context"
	"gs/proto/borrow"
	toolApi "gs/tool/api"
	"log"
	"testing"
	"time"
)

var mc context.Context
var gc borrow.BorrowClient

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
	gc = borrow.NewBorrowClient(conn)

	m.Run()
}

func TestOut(t *testing.T) {
	var books []*borrow.BookInfo
	books = append(books, &borrow.BookInfo{Code: "SN1", Count: 1})
	result, e := gc.OutIn(mc, &borrow.OutInRequest{
		Type:     "借出",
		Username: "测试",
		Books:    books,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestIn(t *testing.T) {
	var books []*borrow.BookInfo
	books = append(books, &borrow.BookInfo{Code: "SN1", Count: 2})
	result, e := gc.OutIn(mc, &borrow.OutInRequest{
		Type:     "归还",
		Username: "测试",
		Books:    books,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}
