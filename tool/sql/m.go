package sql

import (
	"context"
	"fmt"
	"gs/filelog"
	"os"
	"time"

	"gs/tool/env"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// 用户
	TableNameUser = "bs_user"
	// 图书
	TableNameBook = "bs_book"
	// 借还记录
	TableNameBorrowOutIn = "bs_borrow_outin"
	// 用户在借
	TableNameUserBorrow = "bs_user_borrow"
)

var dbPool *pgxpool.Pool

// 创建
func create() {
	// 创建数据表
	_, e := dbPool.Exec(context.Background(), fmt.Sprintf(`
-- 用户
create table if not exists %s (j jsonb);
create unique index if not exists iu_%s_username on %s ((j->'username'));
insert into %s values('{"state":"正常","username":"测试"}') ON CONFLICT ((j->'username')) DO NOTHING;
-- 图书
create table if not exists %s (j jsonb);
create unique index if not exists iu_%s_code on %s ((j->'code'));
-- 借还记录
create table if not exists %s (j jsonb);
-- 用户在借
create table if not exists %s (j jsonb);
create unique index if not exists iu_%s_username on %s ((j->'username'));
`,
		// 用户
		TableNameUser,
		TableNameUser, TableNameUser,
		TableNameUser,
		// 图书
		TableNameBook,
		TableNameBook, TableNameBook,
		// 借还记录
		TableNameBorrowOutIn,
		// 用户在借
		TableNameUserBorrow,
		TableNameUserBorrow, TableNameUserBorrow,
	))

	if e != nil {
		filelog.Warn("创建数据表失败", e.Error())
		os.Exit(1)
	}
}

// 初始化
func Init() {
	// 连接
	postgresUri := env.PostgresUri
	for {
		sqlCtx, sqlCtxCancel := context.WithTimeout(context.Background(), time.Second)
		sdb, e := pgxpool.New(sqlCtx, postgresUri)
		if e != nil {
			filelog.Warn("postgres暂未就绪", e.Error())
			sqlCtxCancel()
			time.Sleep(time.Second * 6)
			continue
		}
		var count int
		e = sdb.QueryRow(sqlCtx, "select count(pid) from pg_stat_activity").Scan(&count)
		if e != nil {
			filelog.Warn("postgres暂未就绪", e.Error())
			sqlCtxCancel()
			time.Sleep(time.Second * 6)
			continue
		}
		sqlCtxCancel()

		dbPool, e = pgxpool.New(context.Background(), postgresUri)
		if e != nil {
			filelog.Warn("postgres暂未就绪", e.Error())
			time.Sleep(time.Second * 6)
			continue
		}
		filelog.Info("postgres就绪")
		break
	}

	// 创建
	create()
}

// 获取
func GetDb() *pgxpool.Pool {
	return dbPool
}
