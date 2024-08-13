package user

import (
	"context"
	"fmt"
	"gs/proto/user"
	"gs/tool"
	toolApi "gs/tool/api"
	toolCache "gs/tool/cache"
	toolSql "gs/tool/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 实现服务
type server struct {
	user.UnimplementedUserServer
}

var sdb *pgxpool.Pool

// Register 注册服务, 传递公共资源
func Register(s grpc.ServiceRegistrar) {
	user.RegisterUserServer(s, &server{})

	sdb = toolSql.GetDb()
}

func (s *server) Add(ctx context.Context, in *user.Info) (*user.Empty, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "需要用户名")
	}
	if tool.ArrayIndex(in.GetState(), []string{"正常", "删除"}) == -1 {
		return nil, status.Errorf(codes.InvalidArgument, "状态无效")
	}

	// 保存入库
	inText, _ := toolApi.ProtoToJson(in)
	ct, e := sdb.Exec(context.Background(), fmt.Sprintf(`insert into %s values('%s');`, toolSql.TableNameUser, inText))
	if e != nil {
		if strings.HasPrefix(e.Error(), "ERROR: duplicate key value") {
			return nil, status.Errorf(codes.InvalidArgument, "用户名重复")
		}

		return nil, status.Errorf(codes.Internal, e.Error())
	}
	if ct.RowsAffected() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "保存失败")
	}

	return &user.Empty{}, nil
}

func (s *server) Change(ctx context.Context, in *user.Info) (*user.Empty, error) {
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "需要用户名")
	}
	if tool.ArrayIndex(in.GetState(), []string{"正常", "删除"}) == -1 {
		return nil, status.Errorf(codes.InvalidArgument, "状态无效")
	}

	// 保存入库
	ct, e := sdb.Exec(context.Background(), fmt.Sprintf(`update %s set j = jsonb_set(j, '{state}', '"%s"') where j->>'username' = '%s';`, toolSql.TableNameUser, in.GetState(), in.GetUsername()))
	if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	}
	if ct.RowsAffected() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "用户名无效")
	}

	return &user.Empty{}, nil
}

func (s *server) Search(ctx context.Context, in *user.SearchRequest) (*user.SearchResponse, error) {
	// 检查数据
	if in.GetPageStart() < 1 || in.GetPageCount() < 1 || in.GetPageCount() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "该页开始行数必须大于1, 该页数量必须在1到100以内")
	}
	sqlWhere := `1 = 1`
	if in.Keyword != "" {
		sqlWhere = fmt.Sprintf(`%s and concat(j->>'username') like '%s'`,
			sqlWhere,
			fmt.Sprint(`%`, in.Keyword, `%`),
		)
	}

	// 构建state查询条件
	if in.GetState() != "" {
		sqlWhere = fmt.Sprintf(`%s and j->'state' ? '%s'`, sqlWhere, in.GetState())
	}

	sqlCount := fmt.Sprintf(`select count(*) from %s as me where %s`, toolSql.TableNameUser, sqlWhere)
	sqlMain := fmt.Sprintf(`select j from %s where %s order by j->>'username'`, toolSql.TableNameUser, sqlWhere)
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
	var infoArray []*user.Info
	for rows.Next() {
		var jsonText string
		e = rows.Scan(&jsonText)
		if e != nil {
			return nil, status.Errorf(codes.Internal, e.Error())
		}
		var info user.Info
		e = toolApi.JsonToProto(jsonText, &info)
		if e != nil {
			return nil, status.Errorf(codes.Internal, e.Error())
		}
		infoArray = append(infoArray, &info)
	}
	result := user.SearchResponse{Count: count, InfoArray: infoArray}
	return &result, nil
}

func (s *server) Auth(ctx context.Context, in *user.AuthRequest) (*user.AuthResponse, error) {
	// 校验数据
	if in.GetUsername() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "需要用户名")
	}
	var jsonText string
	e := sdb.QueryRow(ctx, fmt.Sprintf(`select j from %s where j->>'username' = '%s'`, toolSql.TableNameUser, in.GetUsername())).Scan(&jsonText)
	if e == pgx.ErrNoRows {
		return nil, status.Errorf(codes.InvalidArgument, "用户名无效")
	} else if e != nil {
		return nil, status.Errorf(codes.Internal, e.Error())
	}

	// 更新凭证
	newToken := uuid.New().String()
	toolCache.SetUserToken(in.GetUsername(), newToken)

	// TODO 发出推送

	return &user.AuthResponse{Token: newToken}, nil
}

func (s *server) Exit(ctx context.Context, in *user.Empty) (*user.Empty, error) {
	// 获取上下文
	userId := ctx.Value(toolApi.ContextKeyUserId).(string)

	// TODO 发出推送

	// 删除凭证
	toolCache.DelUserToken(userId)

	return &user.Empty{}, nil
}
