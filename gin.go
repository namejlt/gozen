package gozen

import (
	"context"
	"errors"
	"github.com/gin-gonic/contrib/expvar"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

type OnShutdownF struct {
	f       func(cancel context.CancelFunc)
	timeout time.Duration
}

var (
	onShutdown []OnShutdownF
)

func RegisterOnShutdown(f func(cancel context.CancelFunc), timeout time.Duration) {
	onShutdown = append(onShutdown, OnShutdownF{
		f:       f,
		timeout: timeout,
	})
}

// NewGin 新建gin
func NewGin(param ...string) *gin.Engine {
	/**

	参数param

	第一位 swagger instanceName

	*/
	if ConfigEnvIsDev() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(Recovery())
	if ConfigEnvIsDev() {
		ginpprof.Wrap(r)
		r.Use(gin.Logger())
	}
	routeMonitor(r)
	if configDocsIsExist() {
		var instanceName string
		if len(param) > 0 {
			instanceName = param[0]
		}
		routeSwagger(r, instanceName)
	}
	if configDebugVarsIsExist() {
		routeDebugVars(r)
	}
	return r
}

func ListenHttp(httpPort string, r http.Handler, timeout int, f ...func()) {
	srv := &http.Server{
		Addr:    httpPort,
		Handler: r,
	}
	// 监听端口
	go GoFuncOne(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
			return err
		}
		return nil
	})
	// 注册关闭使用函数
	for _, v := range f {
		srv.RegisterOnShutdown(v)
	}
	// 监听信号
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit
	// 执行on shutdown 函数 - 同步
	for _, v := range onShutdown {
		var wg sync.WaitGroup
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.TODO())
		go v.f(cancel)
		select {
		case <-time.After(v.timeout):
			log.Println("on shutdown timeout:", f)
			LogErrorw(LogNameLogic, "on shutdown timeout",
				"func", f)
			wg.Done()
		case <-ctx.Done():
			wg.Done()
		}
		wg.Wait()
	}
	// 执行shutdown
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func routeMonitor(r *gin.Engine) {
	r.GET("/monitor", func(c *gin.Context) {
		codeVersion := getEnv("CODE_VERSION", "unknown")
		buildTime := getEnv("BUILD_TIME", "unknown")
		systemTime := getSystemTime()
		c.JSON(http.StatusOK, gin.H{
			"code_version": codeVersion,
			"system_time":  systemTime,
			"build_time":   buildTime,
		})
	})
}

func routeSwagger(r *gin.Engine, instanceName string) {
	r.GET("/swagger/*any", gs.WrapHandler(swaggerfiles.Handler,
		gs.InstanceName(configDocsInstanceNameGet(instanceName))),
	)
}

func routeDebugVars(r *gin.Engine) {
	pre := configDebugVarsGet()
	if pre != "" {
		pre = "/" + pre
	}
	r.GET(pre+"/debug/vars", expvar.Handler())
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func getSystemTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
