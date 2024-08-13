package borrow

import (
	"context"
	"fmt"
	"gs/filelog"
	"gs/proto/borrow"
	"gs/tool"
	toolApi "gs/tool/api"
	toolSql "gs/tool/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 实现服务
type server struct {
	borrow.UnimplementedBorrowServer
}

var sdb *pgxpool.Pool

// Register 注册服务, 传递公共资源
func Register(s grpc.ServiceRegistrar) {
	borrow.RegisterBorrowServer(s, &server{})

	sdb = toolSql.GetDb()
}

func (s *server) OutIn(ctx context.Context, in *borrow.OutInRequest) (*borrow.Empty, error) {
	// 基本检查
	if tool.ArrayIndex(in.GetType(), []string{"借出", "归还"}) == -1 {
		return nil, status.Errorf(codes.InvalidArgument, "类型无效")
	}
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "需要用户名")
	}
	if len(in.GetBooks()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "需要图书信息")
	}

	// 获取上下文
	managerId := ctx.Value(toolApi.ContextKeyUserId).(string)

	// 准备事务
	t, e := sdb.BeginTx(ctx, pgx.TxOptions{})
	if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	}
	defer func() {
		if e != nil {
			filelog.Debug("回滚事务", e.Error())
			t.Rollback(ctx)
		} else {
			filelog.Debug("提交事务")
			t.Commit(ctx)
		}
	}()

	// 检查用户名(演示时不作实现)

	// 更新图书数量
	sqlTemplate := `update %s set j = jsonb_set(j, '{borrow_count}', concat('', cast(j->>'borrow_count' as integer) + %d, '')::jsonb) where j->>'code' = '%s' and cast(j->>'borrow_count' as integer) + %d <= cast(j->>'total_count' as integer);`
	if in.GetType() == "归还" {
		sqlTemplate = `update %s set j = jsonb_set(j, '{borrow_count}', concat('', cast(j->>'borrow_count' as integer) - %d, '')::jsonb) where j->>'code' = '%s' and cast(j->>'borrow_count' as integer) - %d >= 0;`
	}
	for _, bookInfo := range in.GetBooks() {
		tag, error := t.Exec(ctx, fmt.Sprintf(sqlTemplate, toolSql.TableNameBook, bookInfo.GetCount(), bookInfo.GetCode(), bookInfo.GetCount()))
		if error != nil {
			e = error
			return nil, status.Errorf(codes.Internal, e.Error())
		}
		if tag.RowsAffected() == 0 {
			if in.GetType() == "借出" {
				e = fmt.Errorf(`图书编码无效或库存数量不足:%s`, bookInfo.GetCode())
			} else {
				e = fmt.Errorf(`图书编码无效或借阅数量异常:%s`, bookInfo.GetCode())
			}
			return nil, status.Errorf(codes.InvalidArgument, e.Error())
		}
	}

	// 保存入库
	in.Id = uuid.New().String()
	in.DateText = time.Now().Format("2006-01-02T15:04:05")
	in.Manager = managerId
	inText, _ := toolApi.ProtoToJson(in)
	tag, error := t.Exec(context.Background(), fmt.Sprintf(`insert into %s values('%s');`, toolSql.TableNameBorrow, inText))
	if error != nil {
		e = error
		return nil, status.Errorf(codes.Internal, e.Error())
	}
	if tag.RowsAffected() == 0 {
		e = fmt.Errorf(`记录没有保存`)
		return nil, status.Errorf(codes.InvalidArgument, "记录没有保存")
	}

	return &borrow.Empty{}, nil
}
