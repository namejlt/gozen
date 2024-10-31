### 模板系统

main 入口文件，初始化服务启动

server 存放服务启动文件，包含http服务、grpc服务、tcp等其他服务

route 存放所有路由，按照app、api、iapi、admin划分

controller 控制器，仅处理参数获取与响应返回

service 包含所有核心逻辑，base_logic基础逻辑，logic直接供controller调用

model 存放所有db映射模型、缓存映射模型、参数模型、响应数据模型等

```
mparam - 入参
mbase - 非业务逻辑固定数据结构
mtransfer - 业务中间数据结构
mmysql - mysql业务结构 - 这里务必要保证注释准确 gorm严格限制
mmongo - mongo业务结构
mapi - 出参
mredis - redis业务结构
mmq - mq业务结构
apim - 调用第三方接口数据结构

m1234 - m开头的业务结构数据

```

dao 包含ao、api、grpc、mongo、mysql、redis、es等所有数据io的操作

script 脚本分为script和daemon两种

proto 存放proto定义文件

grpc 存放pb.go文件和grpc服务实例化文件

middleware 中间件，gin中间件供http服务使用

pconst 常量目录 变量目录

test 存放单元测试

util 工具函数目录

#### tracer

skywalking

http://127.0.0.1:8080/

## swagger

自动生成接口文档

1. 按照swagger要求给接口代码添加声明式注释，具体参照声明式注释格式。
2. 使用swag工具扫描代码自动生成API接口文档数据
3. 使用gin-swagger渲染在线接口文档页面

## 项目启动操作

配置golang环境

设置go环境变量

GOPROXY=https://goproxy.cn,direct;GOSUMDB=off

go install github.com/swaggo/swag/cmd/swag@latest

go mod tidy

make docs

make build

make run
