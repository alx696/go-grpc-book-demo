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

func TestOut1(t *testing.T) {
	var books []*borrow.BookInfo
	books = append(books, &borrow.BookInfo{Code: "SN1", Count: 1})
	result, e := gc.OutIn(mc, &borrow.OutInInfo{
		Type:  "借出",
		Books: books,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestOut2(t *testing.T) {
	var books []*borrow.BookInfo
	books = append(books, &borrow.BookInfo{Code: "SN2", Count: 3})
	result, e := gc.OutIn(mc, &borrow.OutInInfo{
		Type:  "借出",
		Books: books,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestIn1(t *testing.T) {
	var books []*borrow.BookInfo
	books = append(books, &borrow.BookInfo{Code: "SN1", Count: 1})
	result, e := gc.OutIn(mc, &borrow.OutInInfo{
		Type:  "归还",
		Books: books,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestIn2(t *testing.T) {
	var books []*borrow.BookInfo
	books = append(books, &borrow.BookInfo{Code: "SN2", Count: 2})
	result, e := gc.OutIn(mc, &borrow.OutInInfo{
		Type:  "归还",
		Books: books,
	})
	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}

func TestQueryUserBorrow(t *testing.T) {
	result, e := gc.QueryUserBorrow(mc, &borrow.Empty{})

	if e != nil {
		t.Fatal(e)
	}
	t.Log(result)
}
