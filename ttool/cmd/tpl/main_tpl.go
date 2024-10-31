package tpl

var (
	MainDirName   = ""
	MainFilesName = []string{
		".gitignore",
		".gitlab-ci.yml",
		"go.mod",
		"host.go",
		"main.go",
		"Makefile",
		"README.md",
	}
	MainFilesContent = []string{
		mainGitignore,
		mainGitlabCiYml,
		mainGoMod,
		mainHostGo,
		mainMainGo,
		mainMakefile,
		mainReadmeMd,
	}
)

var (
	mainGitignore = `/.idea
bin/
pkg/
{{.Name}}
dist
`
	mainGitlabCiYml = `variables:
  SONAR_URL: "http://127.0.0.1:9000"
  SONAR_LOGIN: loginname
  SONAR_PROJECT: {{.Name}}

stages:
  - test
  - build
  - deploy
sonarqube-test:
  stage: test
  tags:
    - demo
  script:
    - sonar-scanner -Dsonar.projectKey=$SONAR_PROJECT -Dsonar.host.url=$SONAR_URL -Dsonar.login=$SONAR_LOGIN -Dsonar.sources=. -Dsonar.analysis.CI_COMMIT_REF_NAME=$CI_COMMIT_REF_NAME -Dsonar.analysis.GITLAB_USER_EMAIL=$GITLAB_USER_EMAIL -Dsonar.analysis.GITLAB_USER_NAME=$GITLAB_USER_NAME -Dsonar.analysis.CI_PROJECT_PATH=$CI_PROJECT_PATH

go_package:
  image: golang
  stage: build
  only:
    - master
  tags:
    - demo
  script:
    - echo "build package."
empty-release:
  stage: deploy
  only:
    - master
  tags:
    - demo
  script:
    - echo "empty release."
`
	mainGoMod = `module {{.Name}}

go 1.22.0


toolchain go1.22.1

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/gin-gonic/gin v1.10.0
	github.com/namejlt/gozen v0.0.0-20241031073332-f70af9f9bb5f
	github.com/speps/go-hashids v2.0.0+incompatible
	github.com/swaggo/swag v1.16.4
	github.com/urfave/cli v1.22.16
	go.mongodb.org/mongo-driver v1.17.1
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/net v0.30.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
	gorm.io/gorm v1.25.12
	skywalking.apache.org/repo/goapi v0.0.0-20241023080050-2514649a8007
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/DeanThompson/ginpprof v0.0.0-20201112072838-007b1e56b2e1 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/SkyAPM/go2sky v1.5.0 // indirect
	github.com/alibabacloud-go/darabonba-array v0.1.0 // indirect
	github.com/alibabacloud-go/darabonba-encode-util v0.0.2 // indirect
	github.com/alibabacloud-go/darabonba-map v0.0.2 // indirect
	github.com/alibabacloud-go/darabonba-string v1.0.2 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/alibabacloud-go/openapi-util v0.1.1 // indirect
	github.com/alibabacloud-go/tea v1.2.2 // indirect
	github.com/alibabacloud-go/tea-utils/v2 v2.0.7 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.44 // indirect
	github.com/aliyun/alibabacloud-dkms-gcs-go-sdk v0.5.1 // indirect
	github.com/aliyun/alibabacloud-dkms-transfer-go-sdk v0.1.9 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/bytedance/sonic v1.12.3 // indirect
	github.com/bytedance/sonic/loader v0.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/didip/tollbooth v4.0.2+incompatible // indirect
	github.com/gabriel-vasile/mimetype v1.4.6 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/contrib v0.0.0-20240508051311-c1c6bf0061b0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-stomp/stomp v2.1.4+incompatible // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jolestar/go-commons-pool v2.0.0+incompatible // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nacos-group/nacos-sdk-go/v2 v2.2.7 // indirect
	github.com/olivere/elastic v6.2.37+incompatible // indirect
	github.com/olivere/elastic/v7 v7.0.32 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.60.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	github.com/swaggo/gin-swagger v1.6.0 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/arch v0.11.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/olivere/elastic.v3 v3.0.75 // indirect
	gopkg.in/olivere/elastic.v5 v5.0.86 // indirect
	gopkg.in/olivere/elastic.v6 v6.2.37 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.5.7 // indirect
)


replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20241021214115-324edc3d5d38
`
	mainHostGo = `package main

import (
	"github.com/namejlt/gozen"
	"io/fs"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	//替换main.go 里面 @host 文案
	var toR string
	hostTag := "// @host"

	f := "./main.go"

	h := hostTag + " " + gozen.ConfigAppGetString("Host", "localhost:8888")

	sAll, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	sAllStr := string(sAll)

	//查找匹配字符串
	re := regexp.MustCompile("// @host[\\s][\\w\\.:]*")

	sArr := re.FindAllString(sAllStr, 1)

	if len(sArr) != 1 {
		return
	}
	if strings.Trim(sArr[0], "\n") == hostTag {
		toR = hostTag
	} else {
		toR = sArr[0]
	}

	sAllStrTodo := strings.Replace(sAllStr, toR, h, 1)

	err = ioutil.WriteFile(f, []byte(sAllStrTodo), fs.ModePerm)
	if err != nil {
		panic(err)
	}
}
`
	mainMainGo = `package main

import (
	"os"
	"{{.Name}}/script"
	"{{.Name}}/server"

	"github.com/namejlt/gozen"
	"github.com/urfave/cli"
	_ "go.uber.org/automaxprocs"
)

// @title {{.Name}}系统
// @version 1.0
// @description 这是一个{{.Name}}系统，用于提供框架项目相关开发示例
// @termsOfService https://tynam.com

// @contact.name tynam
// @contact.url https://tynam.com
// @contact.email jiaongtian@163.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

//go:generate go run host.go
// @host localhost:8888
// @BasePath /
func main() {
	defer gozen.LogSync()
	app := cli.NewApp()
	app.Name = "{{.Name}} script tool"
	app.Usage = "run scripts!"
	app.Version = "0.0.1"
	app.Author = "anonymous"
	app.Commands = script.Commands()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Value: "http",
			Usage: "run server type:  http, grpc",
		},
	}
	app.Action = func(c *cli.Context) error {
		println("RunHttp Server.")
		serverType := c.String("server")
		switch serverType {
		case "http":
			server.RunHttp()
		case "grpc":
			server.RunGRpc()
		default:
			server.RunHttp()
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		panic("app run error:" + err.Error())
	}
}
`
	mainMakefile = `# Go parameters
GOCMD=GO111MODULE=on CGO_ENABLED=1 go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGENERATE=$(GOCMD) generate
GOMODTIDY=$(GOCMD) mod tidy

all: build run

docs: cleandocs initdocs

build:
	rm -rf dist/
	mkdir -p dist/configs
	cp -rf configs dist/
	$(GOMODTIDY)
	$(GOGENERATE)
	$(GOBUILD) -o dist/{{.Name}} main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf dist/

cleandocs:
	rm -rf docs/

run:
	dist/{{.Name}}

stop:
	pkill -f dist/{{.Name}}

fetch:
	ps aux | grep {{.Name}}

initdocs:
	swag init
`
	mainReadmeMd = `### {{.Name}}系统

main 入口文件，初始化服务启动

server 存放服务启动文件，包含http服务、grpc服务、tcp等其他服务

route 存放所有路由，按照app、api、iapi、admin划分

controller 控制器，仅处理参数获取与响应返回

service 包含所有核心逻辑，base_logic基础逻辑，logic直接供controller调用

model 存放所有db映射模型、缓存映射模型、参数模型、响应数据模型等


mparam - 入参
mbase - 非业务逻辑固定数据结构
mtransfer - 业务中间数据结构
mmysql - mysql业务结构
mmongo - mongo业务结构
mapi - 出参
mredis - redis业务结构
mmq - mq业务结构
apim - 调用第三方接口数据结构

m1234 - m开头的业务结构数据



dao 包含ao、api、grpc、mongo、mysql、redis、es等所有数据io的操作

script 脚本分为script和daemon两种

proto 存放proto定义文件

grpc 存放pb.go文件和grpc服务实例化文件

middleware 中间件，gin中间件供http服务使用

pconst 常量目录 变量目录

test 存放单元测试

util 工具函数目录

#### tracer

http://skywalking-service:8080/

## swagger

自动生成接口文档

1. 按照swagger要求给接口代码添加声明式注释，具体参照声明式注释格式。
2. 使用swag工具扫描代码自动生成API接口文档数据
3. 使用gin-swagger渲染在线接口文档页面


## 项目启动操作

1. 配置golang环境 下载golang包并解压
2. 设置go环境变量  GOPROXY=https://goproxy.cn,direct;GOSUMDB=off
4. go get -u github.com/swaggo/swag/cmd/swag 并build 设置变量 保证可以执行swag命令
5. go mod tidy
6. make docs
7. make build
8. make run

## 操作实践


1. 查看文档 http://localhost:8888/swagger/any/index.html
2. 查看链路日志 skywalking服务
3. 查看日志，日志走elk，通过kibana查看
4. 默认服务 http://localhost:8888/
`
)
