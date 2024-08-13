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

func (s *server) OutIn(ctx context.Context, in *borrow.OutInInfo) (*borrow.Empty, error) {
	// 基本检查
	if tool.ArrayIndex(in.GetType(), []string{"借出", "归还"}) == -1 {
		return nil, status.Errorf(codes.InvalidArgument, "类型无效")
	}
	if len(in.GetBooks()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "需要图书信息")
	}

	// 上下文中获取用户名
	username := ctx.Value(toolApi.ContextKeyUserId).(string)

	// // 查询用户在借
	// userBorrowTx, userBorrowTxError := sdb.BeginTx(ctx, pgx.TxOptions{})
	// if userBorrowTxError != nil {
	// 	return nil, status.Errorf(codes.Internal, userBorrowTxError.Error())
	// }
	// defer func() {
	// 	if userBorrowTxError != nil {
	// 		filelog.Debug("回滚查询用户在借事务", userBorrowTxError.Error())
	// 		userBorrowTx.Rollback(ctx)
	// 	} else {
	// 		filelog.Debug("提交查询用户在借事务")
	// 		userBorrowTx.Commit(ctx)
	// 	}
	// }()
	// var userBorrowJsonText string
	// e := userBorrowTx.QueryRow(ctx, fmt.Sprintf(`select j from %s where j->>'username' = '%s' FOR UPDATE;`, toolSql.TableNameUserBorrow, username)).Scan(&userBorrowJsonText)
	// if e != nil && e != pgx.ErrNoRows {
	// 	userBorrowTxError = e
	// 	return nil, status.Errorf(codes.Internal, e.Error())
	// }

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

	// 查询用户在借
	var diffBooks []*borrow.BookInfo
	var userBorrow borrow.UserBorrow
	var userBorrowJsonText string
	e = t.QueryRow(ctx, fmt.Sprintf(`select j from %s where j->>'username' = '%s' FOR UPDATE;`, toolSql.TableNameUserBorrow, username)).Scan(&userBorrowJsonText)
	if e == pgx.ErrNoRows {
		e = nil
		if in.GetType() == "借出" {
			diffBooks = in.GetBooks()
		}
	} else if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	} else {
		e = toolApi.JsonToProto(userBorrowJsonText, &userBorrow)
		if e != nil {
			return nil, status.Errorf(codes.Internal, e.Error())
		}

		for _, bookInfoNow := range in.GetBooks() {
			var existsCount int32
			for _, bookInfoExists := range userBorrow.GetBooks() {
				if bookInfoExists.GetCode() == bookInfoNow.GetCode() {
					existsCount = bookInfoExists.GetCount()
					continue
				}
			}

			var diffCount int32
			if in.GetType() == "借出" {
				diffCount = bookInfoNow.GetCount() + existsCount
			} else {
				diffCount = existsCount - bookInfoNow.GetCount()
			}
			if diffCount > 0 {
				diffBooks = append(diffBooks, &borrow.BookInfo{Code: bookInfoNow.GetCode(), Count: diffCount})
			}
		}
	}
	// 更新用户在借
	if len(diffBooks) > 0 {
		userBorrow = borrow.UserBorrow{Username: username, Books: diffBooks}
		userBorrowJsonText, _ = toolApi.ProtoToJson(&userBorrow)
		tag, error := t.Exec(ctx, fmt.Sprintf(`insert into %s values('%s') ON CONFLICT ((j->'username')) DO UPDATE SET j = EXCLUDED.j;`, toolSql.TableNameUserBorrow, userBorrowJsonText))
		if error != nil {
			e = error
			return nil, status.Errorf(codes.Internal, e.Error())
		}
		if tag.RowsAffected() == 0 {
			e = fmt.Errorf(`用户在借没有保存`)
			return nil, status.Errorf(codes.InvalidArgument, "用户在借没有保存")
		}
	} else {
		_, error := t.Exec(ctx, fmt.Sprintf(`delete from %s where j->>'username' = '%s';`, toolSql.TableNameUserBorrow, username))
		if error != nil {
			e = error
			return nil, status.Errorf(codes.Internal, e.Error())
		}
	}

	// 保存入库
	in.Id = uuid.New().String()
	in.DateText = time.Now().Format("2006-01-02T15:04:05")
	in.Username = username
	inText, _ := toolApi.ProtoToJson(in)
	tag, error := t.Exec(ctx, fmt.Sprintf(`insert into %s values('%s');`, toolSql.TableNameBorrowOutIn, inText))
	if error != nil {
		e = error
		return nil, status.Errorf(codes.Internal, e.Error())
	}
	if tag.RowsAffected() == 0 {
		e = fmt.Errorf(`借还记录没有保存`)
		return nil, status.Errorf(codes.InvalidArgument, "借还记录没有保存")
	}

	return &borrow.Empty{}, nil
}
