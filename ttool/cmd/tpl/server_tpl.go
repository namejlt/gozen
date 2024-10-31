package tpl

var (
	ServerDirName   = "server"
	ServerFilesName = []string{
		"grpc.go",
		"http.go",
		"init.go",
	}
	ServerFilesContent = []string{
		ServerGRpcGo,
		ServerHttpGo,
		ServerInitGo,
	}
)

var (
	ServerGRpcGo = `package server

import (
	"fmt"
	"log"
	"net"
	"{{.Name}}/grpc/test"
	"{{.Name}}/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var grpcPort = ":9999"

func RunGRpc() {
	//环境初始化
	configRuntime()
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("success", grpcPort)
	var options []grpc.ServerOption
	middleware.Register(&options)
	s := grpc.NewServer(options...)
	//定义路由服务
	test.RegisterTestServer(s, &test.GTest{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("grpc run ok", grpcPort)
}
`
	ServerHttpGo = `package server

import (
	"fmt"
	"{{.Name}}/middleware"
	"{{.Name}}/route"

	_ "{{.Name}}/docs"

	"github.com/namejlt/gozen"
)

var httpPort = ":8888"

func RunHttp() {
	//环境初始化
	configRuntime()
	gozen.InitSkyWalking()
	//开启服务
	StartListening()
}

func StartListening() {
	r := gozen.NewGin()
	r.Use(middleware.Monitor(), gozen.MiddlewareHttp())

	route.RouteIApi(r)
	route.RouteApi(r)
	route.RouteApp(r)
	route.RouteAdmin(r)
	route.RouteHome(r)

	fmt.Println("start", httpPort)
	gozen.ListenHttp(httpPort, r, 10, gozen.CloseSkyWalking)
}
`
	ServerInitGo = `package server

import (
	"fmt"
	"runtime"
	"time"
)

func configRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	now := time.Now().String()
	fmt.Printf("Running time is %s\n", now)
	fmt.Printf("Running with %d CPUs\n", nuCPU)
}

`
)
