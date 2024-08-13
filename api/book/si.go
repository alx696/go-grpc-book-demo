package book

import (
	"context"
	"fmt"
	"gs/proto/book"
	"gs/tool"
	toolApi "gs/tool/api"
	toolSql "gs/tool/sql"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 实现服务
type server struct {
	book.UnimplementedBookServer
}

var sdb *pgxpool.Pool

// Register 注册服务, 传递公共资源
func Register(s grpc.ServiceRegistrar) {
	book.RegisterBookServer(s, &server{})

	sdb = toolSql.GetDb()
}

func (s *server) Add(ctx context.Context, in *book.Info) (*book.Empty, error) {
	if in.GetCode() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "需要图书编码")
	}
	if tool.ArrayIndex(in.GetState(), []string{"正常", "删除"}) == -1 {
		return nil, status.Errorf(codes.InvalidArgument, "状态无效")
	}

	// 保存入库
	inText, _ := toolApi.ProtoToJson(in)
	ct, e := sdb.Exec(context.Background(), fmt.Sprintf(`insert into %s values('%s');`, toolSql.TableNameBook, inText))
	if e != nil {
		if strings.HasPrefix(e.Error(), "ERROR: duplicate key value") {
			return nil, status.Errorf(codes.InvalidArgument, "图书编码重复")
		}

		return nil, status.Errorf(codes.Internal, e.Error())
	}
	if ct.RowsAffected() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "保存失败")
	}

	return &book.Empty{}, nil
}

func (s *server) Change(ctx context.Context, in *book.Info) (*book.Empty, error) {
	if in.GetCode() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "需要图书编码")
	}
	if tool.ArrayIndex(in.GetState(), []string{"正常", "删除"}) == -1 {
		return nil, status.Errorf(codes.InvalidArgument, "状态无效")
	}

	// 保存入库
	ct, e := sdb.Exec(context.Background(), fmt.Sprintf(`update %s set j = j || '{"name": "%s", "total_count": %d, "state": "%s"}' where j->>'code' = '%s';`, toolSql.TableNameBook, in.GetName(), in.GetTotalCount(), in.GetState(), in.GetCode()))
	if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	}
	if ct.RowsAffected() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "图书编码无效")
	}

	return &book.Empty{}, nil
}

func (s *server) Search(ctx context.Context, in *book.SearchRequest) (*book.SearchResponse, error) {
	// 检查数据
	if in.GetPageStart() < 1 || in.GetPageCount() < 1 || in.GetPageCount() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "该页开始行数必须大于1, 该页数量必须在1到100以内")
	}
	sqlWhere := `1 = 1`
	if in.Keyword != "" {
		sqlWhere = fmt.Sprintf(`%s and concat(j->>'code', j->>'name') like '%s'`,
			sqlWhere,
			fmt.Sprint(`%`, in.Keyword, `%`),
		)
	}

	sqlCount := fmt.Sprintf(`select count(*) from %s as me where %s`, toolSql.TableNameBook, sqlWhere)
	sqlMain := fmt.Sprintf(`select j from %s where %s order by j->>'code'`, toolSql.TableNameBook, sqlWhere)
	sqlPage := fmt.Sprintf(`offset %v limit %v`, in.PageStart-1, in.PageCount)
	var sqlFull string
	if in.PageStart > 0 {
		sqlFull = fmt.Sprint(sqlMain, " ", sqlPage)
	} else {
		sqlFull = sqlMain
	}

	var count int32
	e := sdb.QueryRow(ctx, sqlCount).Scan(&count)
	if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	}

	rows, e := sdb.Query(ctx, sqlFull)
	if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	}
	var infoArray []*book.Info
	for rows.Next() {
		var jsonText string
		e = rows.Scan(&jsonText)
		if e != nil {
			return nil, status.Errorf(codes.Internal, e.Error())
		}
		var info book.Info
		e = toolApi.JsonToProto(jsonText, &info)
		if e != nil {
			return nil, status.Errorf(codes.Internal, e.Error())
		}
		infoArray = append(infoArray, &info)
	}
	result := book.SearchResponse{Count: count, InfoArray: infoArray}
	return &result, nil
}
