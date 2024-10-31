package gozen

import (
	"context"
	"errors"
	"github.com/namejlt/gozen/pool"
	"log"
)

// 连接GRPC服务并自建负载均衡
type DaoGRPC struct {
	ServerName string
	Ctx        context.Context
}

var (
	gRpcPool map[string]pool.Pool
)

// 初始化连接池
func InitGRpcPool() {
	conf := poolConfigMap
	if conf != nil {
		gRpcPool = make(map[string]pool.Pool, len(poolConfigMap.Configs))
		for k, v := range poolConfigMap.Configs {
			opt := pool.DefaultOptions
			opt.MaxIdle = v.MaxIdle
			opt.MaxActive = v.MaxActive
			opt.MaxConcurrentStreams = v.MaxConcurrentStreams
			opt.Reuse = v.Reuse
			p, err := pool.New(v.Address, opt)
			if err != nil {
				log.Fatalf("failed to new pool: %v", err)
			}
			gRpcPool[k] = p
		}
	}
}

func CloseGRpcPool() {
	for _, v := range gRpcPool {
		_ = v.Close()
	}
}

func daoGRPCGetPool(serverName string) (p pool.Pool, err error) {
	poolName := "grpc-" + serverName
	var ok bool
	p, ok = gRpcPool[poolName]
	if !ok {
		err = errors.New("grpc server not exist")
		return
	}
	return
}

// 获取连接
func (p *DaoGRPC) GetConn() (conn pool.Conn, err error) {
	poolT, err := daoGRPCGetPool(p.ServerName)
	if err != nil {
		return
	}

	conn, err = poolT.Get()

	return
}

func (p *DaoGRPC) CloseConn(conn pool.Conn) (err error) {
	err = conn.Close()
	return
}
