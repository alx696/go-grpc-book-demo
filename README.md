# 开发

## 需求

原始需求:

```
功能需求
只实现 api 接口即可， 使用 接口测试工具 进行 测试
1. 实现 book 增删改查， 只需包含， 名称， 数量， 图书编码（唯一）
2. 实现 用户 增删改查 ， 只需要 包含 用户名
3. 实现 用户 借用图书， 还书，并查询用户借阅的图书
技术要求
1. 图书借阅 正确使用 事务， 保证 数据一致性
2. 实现 所有接口 单元测试
3. 项目组织良好
4. 代码commit 记录完整，按照功能点提交commit
产物交付
1. 录制程序接口调试视频
a. 演示 所有功能需求
2. 讲解项目结构，以及设计思路
3. 代码推送到gitee, 提供源代码链接
```

整理之后:

* 用户:增改删查,登录,退出.用户名要做唯一性校验
* 图书:增改删查.图书编码要做唯一性校验
* 借还:借书,还书,查询用户在借图书(不实现借阅记录查询).

关键点:

* 唯一性校验:图书编码,用户名
* 借出数量不能超过库存数量, 归还数量不能超过借出数量
* 为了提升查询效率, 每次借还操作时更新图书借出数量, 后续借出时只需查询图书信息即可
* 为了提升查询效率, 每次借还操作时更新用户在借数据, 后续查询时无需全部扫描借还记录

## 数据库

postgres:

```
docker run -d --restart=always \
  --shm-size="4g" \
  -p 5432:5432 \
  -v "${PWD}/postgres":/data \
  -e PGDATA=/data -e TZ=Asia/Shanghai -e POSTGRES_PASSWORD=postgres \
  --name "postgres" registry.cn-shanghai.aliyuncs.com/xm69/postgres:16 \
  -c "max_connections=1000" \
  -c "idle_session_timeout=60000" \
  -c "idle_in_transaction_session_timeout=120000" \
  -c "max_wal_size=4GB" \
  -c "shared_buffers=4GB" \
  -c "work_mem=64MB" \
  -c "maintenance_work_mem=2GB" \
  -c "checkpoint_completion_target=0.9" \
  -c "random_page_cost=1.1" \
  -c "effective_io_concurrency=200" \
  -c "track_io_timing=on" \
  -c "default_statistics_target=500" \
  -c "jit=off" \
  -c "log_statement=all" \
  -c "log_min_duration_statement=1000" \
  -c "log_line_prefix='%m [%p] [%r] '"
```

调试工具:

```
docker run -d --restart=always \
  -p 5433:80 \
  -e "PGADMIN_DEFAULT_EMAIL=p@g.cn" \
  -e "PGADMIN_DEFAULT_PASSWORD=p" \
  --name "postgres-pgadmin" dpage/pgadmin4:latest
```

## 缓存

生产环境一般使用Redis, 作演示直接使用map.

> 使用map不方便, 服务一重启用户凭证就丢失了.

## 运行

> 注意: 需要到 proto 文件夹中(README.md中有说明)生成协议代码! 使用 Visual Studio Code 开发.

### 创建环境变量文件 `.env` :

```
# 系统id(开发在后面添加自己名字拼音首字母的缩写)
BS_SERVICE_SYSTEM_ID=service_bs_dev

# gRPC端口
BS_SERVICE_GRPC_PORT=20000

# postgres
POSTGRES_URI="postgres://postgres:postgres@172.17.0.1:5432/postgres?sslmode=disable&pool_max_conns=100&pool_max_conn_lifetime=1m&pool_max_conn_idle_time=1m&application_name=service_bs"
```

### 创建测试环境变量文件 `.env-test` :

```
# 服务地址
GRPC_TARGET=localhost:20000

# 服务是否启用TLS
GRPC_TLS=false

# 用户凭证
TOKEN=d6449e41-e039-4458-8ab4-b47516aeacb1
```

默认添加了用户名为 `测试` 的用户, 首先需要调用 `/user.User/Auth` 登录, 将返回的token设置到 `.env-test` 中的用户凭证, 然后执行其它测试用例.

没有实现完全的自动化测试, 每个接口均实现了测试单例, 直接运行或调整参数后运行即可.

