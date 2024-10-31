package gozen

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func ListenGrpc(grpcPort string, srv *grpc.Server) {
	// 监听端口
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Server Listen", grpcPort)

	// 启动服务
	go GoFuncOne(func() error {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		return nil
	})

	log.Println("Server run")

	// 监听信号
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-quit

	// 退出服务
	log.Println("Shutdown Server ...")
	srv.GracefulStop()
	log.Println("Server exiting")
}
