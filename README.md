# gozen

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/namejlt/gozen)

#### 特性

~~~~
快速创建项目
封装支持常用中间件:mysql、redis、mongodb、es
分层运行:http、grpc、script
链路跟踪使用skywalking

~~~~

#### 模块

~~~~
config:配置
util:工具
dao:数据访问
model:数据模型
gin配置路由

~~~~

#### 生命周期说明

~~~~

启动

0、扫描配置检测格式并配置写进内存
1、针对配置进行初始化操作：db连接池、grpc连接池、cache连接池或其他服务检测可用性
2、启动http或grpc服务

关闭
0、信号通知进程关闭，进入关闭流程http、grpc走shutdown流程
1、shutdown中针对配置之前初始化的服务进行主动close操作

~~~~

#### redis

~~~~

封装 github.com/go-redis/redis

redis 主从
配置多个主节点，

redis cluster

包含侵入非通用redis地址获取代码


~~~~


