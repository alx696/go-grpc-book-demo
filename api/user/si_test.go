package user

import (
	"context"
	"fmt"
	"gs/proto/user"
	toolApi "gs/tool/api"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var mc context.Context
var gc user.UserClient
var token string

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

func TestAdd(t *testing.T) {
	// 基本校验
	// 用户名
	_, e := gc.Add(mc, &user.Info{})
	gs, gsOk := status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.InvalidArgument {
		t.Fatal("基本校验:用户名", gs.Code(), gs.Message())
	}
	// 状态
	_, e = gc.Add(mc, &user.Info{
		Username: "测试",
		State:    "异常",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.InvalidArgument {
		t.Fatal("基本校验:状态", gs.Code(), gs.Message())
	}

	// 用户名重复
	_, e = gc.Add(mc, &user.Info{
		Username: "测试",
		State:    "正常",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.InvalidArgument {
		t.Fatal("用户名重复", gs.Code(), gs.Message())
	}

	// 正常添加
	_, e = gc.Add(mc, &user.Info{
		Username: fmt.Sprint("测试添加-", uuid.New().String()),
		State:    "正常",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK {
		t.Fatal("正常添加", gs.Code(), gs.Message())
	}
}

func TestChange(t *testing.T) {
	// 用户名无效
	_, e := gc.Change(mc, &user.Info{
		Username: "不存在",
		State:    "删除",
	})
	gs, gsOk := status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.InvalidArgument {
		t.Fatal("用户名无效", gs.Code(), gs.Message())
	}

	username := fmt.Sprint("测试修改状态-", uuid.New().String())

	// 添加用户
	_, e = gc.Add(mc, &user.Info{
		Username: username,
		State:    "正常",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK {
		t.Fatal("添加用户", gs.Code(), gs.Message())
	}

	// 正常修改
	_, e = gc.Change(mc, &user.Info{
		Username: username,
		State:    "删除",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK {
		t.Fatal("正常修改", gs.Code(), gs.Message())
	}
}

func TestSearch(t *testing.T) {
	// 起始页码
	_, e := gc.Search(mc, &user.SearchRequest{
		PageStart: 0,
		PageCount: 10,
	})
	gs, gsOk := status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.InvalidArgument {
		t.Fatal("起始页码", gs.Code(), gs.Message())
	}
	// 每页条数
	_, e = gc.Search(mc, &user.SearchRequest{
		PageStart: 1,
		PageCount: 101,
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.InvalidArgument {
		t.Fatal("每页条数", gs.Code(), gs.Message())
	}

	// 关键字
	result, e := gc.Search(mc, &user.SearchRequest{
		PageStart: 1,
		PageCount: 10,
		Keyword:   "测试",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK || result.GetCount() == 0 {
		t.Fatal("关键字", gs.Code(), gs.Message())
	}
	t.Log("关键字测试数量", result.GetCount())

	// 状态
	result, e = gc.Search(mc, &user.SearchRequest{
		PageStart: 1,
		PageCount: 10,
		State:     "正常",
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK || result.GetCount() == 0 {
		t.Fatal("状态", gs.Code(), gs.Message())
	}
	t.Log("正常状态数量", result.GetCount())
}

func TestAuth(t *testing.T) {
	username := fmt.Sprint("测试登录-", uuid.New().String())

	// 添加用户
	_, e := gc.Add(mc, &user.Info{
		Username: username,
		State:    "正常",
	})
	gs, gsOk := status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK {
		t.Fatal("添加用户", gs.Code(), gs.Message())
	}

	// 验证
	result, e := gc.Auth(mc, &user.AuthRequest{
		Username: username,
	})
	gs, gsOk = status.FromError(e)
	if !gsOk {
		t.Fatal(e)
	}
	if gs.Code() != codes.OK {
		t.Fatal("验证", gs.Code(), gs.Message())
	}
	token = result.GetToken()
	t.Log("获得凭证", token)
}

// // 不执行, 以免在批量测试时影响其它测试.
// func TestExit(t *testing.T) {
// 	_, e := gc.Exit(mc, &user.Empty{})
// 	if e != nil {
// 		t.Fatal(e)
// 	}
// }
