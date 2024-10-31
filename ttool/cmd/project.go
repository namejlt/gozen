package cmd

import (
	"fmt"
	"github.com/namejlt/gozen/ttool/cmd/tpl"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"text/template"
)

/**

工程管理

*/

func init() {
	projectCmd.Flags().StringP("name", "n", "", "project name")
	projectCmd.Flags().StringP("action", "a", "init", "project action")
}

const (
	ProjectCmdActionInit = "init" //项目初始化
)

var projectCmd = &cobra.Command{
	Use:   "project [-n name] [-a action]",
	Short: "project manage",
	Long:  `project init less, project init all`,
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		if name == "" {
			fmt.Fprintln(os.Stderr, "project name is null")
			return
		}
		action, err := cmd.Flags().GetString("action")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		switch action {
		case ProjectCmdActionInit:
			data := ProjectData{
				Name: name,
			}
			initProject(data)
		default:
			fmt.Fprintln(os.Stderr, "action is valid")
		}
		return
	},
}

type ProjectData struct {
	Name string `json:"name"`
}

func initProject(data ProjectData) {
	fmt.Println("start init project", data.Name)
	initConfigs(data)
	initDocs(data)
	initMain(data)
	initProto(data)
	initServer(data)
	initRoute(data)
	initController(data)
	initDao(data)
	initModel(data)
	initPConst(data)
	initScript(data)
	initService(data)
	initMiddleware(data)
	initGRpc(data)
	fmt.Println("end init project", data.Name)
}

func initProjectPath(path string) {
	_ = os.MkdirAll(path, fs.ModePerm)
}

func getProjectPath(data ProjectData) string {
	return "./" + data.Name
}

func initModule(data ProjectData, module string, dir []string, fileName [][]string, fileContent [][]string) {
	fmt.Println(module + " start init")

	for dk, dv := range dir {
		path := getProjectPath(data) + "/" + dv
		initProjectPath(path)

		//生成文件
		for k, fileName := range fileName[dk] {
			fileF, err := os.Create(path + "/" + fileName)
			if err != nil {
				fmt.Fprintln(os.Stderr, "os.Create", err)
				return
			}
			tmpl, err := template.New(ProjectCmdActionInit).Parse(fileContent[dk][k])
			if err != nil {
				fmt.Fprintln(os.Stderr, "template.New.Parse", err)
				return
			}
			err = tmpl.Execute(fileF, data)
			if err != nil {
				fmt.Fprintln(os.Stderr, "tmpl.Execute", err)
				return
			}
			fileF.Close()
		}
	}

	fmt.Println(module + " end init")
}

func initModuleNotTpl(data ProjectData, module string, dir []string, fileName [][]string, fileContent [][]string) {
	fmt.Println(module + " start init")

	for dk, dv := range dir {
		path := getProjectPath(data) + "/" + dv
		initProjectPath(path)

		//生成文件
		for k, fileName := range fileName[dk] {
			fileF, err := os.Create(path + "/" + fileName)
			if err != nil {
				fmt.Fprintln(os.Stderr, "os.Create", err)
				return
			}
			_, err = fileF.WriteString(fileContent[dk][k])
			if err != nil {
				fmt.Fprintln(os.Stderr, "fileF.WriteString", err)
				return
			}
			fileF.Close()
		}
	}

	fmt.Println(module + " end init")
}

func initConfigs(data ProjectData) {
	initModule(data, "configs", []string{tpl.ConfigsDirName}, [][]string{tpl.ConfigsFilesName}, [][]string{tpl.ConfigsFilesContent})
}

func initDocs(data ProjectData) {
	initModuleNotTpl(data, "docs", []string{tpl.DocsDirName}, [][]string{tpl.DocsFilesName}, [][]string{tpl.DocsFilesContent})
}

func initController(data ProjectData) {
	initModule(data, "controller",
		[]string{
			tpl.ControllerDirName,
			tpl.ControllerV1DirName,
		},
		[][]string{
			tpl.ControllerFilesName,
			tpl.ControllerV1FilesName,
		},
		[][]string{
			tpl.ControllerFilesContent,
			tpl.ControllerV1FilesContent,
		},
	)
}

func initDao(data ProjectData) {
	initModule(data, "dao",
		[]string{
			tpl.DaoApiDirName,
			tpl.DaoMongoDirName,
			tpl.DaoMysqlDirName,
			tpl.DaoRedisDirName,
		},
		[][]string{
			tpl.DaoApiFilesName,
			tpl.DaoMongoFilesName,
			tpl.DaoMysqlFilesName,
			tpl.DaoRedisFilesName,
		},
		[][]string{
			tpl.DaoApiFilesContent,
			tpl.DaoMongoFilesContent,
			tpl.DaoMysqlFilesContent,
			tpl.DaoRedisFilesContent,
		},
	)
}

func initMiddleware(data ProjectData) {
	initModule(data, "middleware", []string{tpl.MiddlewareDirName}, [][]string{tpl.MiddlewareFilesName}, [][]string{tpl.MiddlewareFilesContent})
}

func initModel(data ProjectData) {
	initModule(data, "model",
		[]string{
			tpl.ModelApiMDirName,
			tpl.ModelMApiDirName,
			tpl.ModelMBaseDirName,
			tpl.ModelMMongoDirName,
			tpl.ModelMMysqlDirName,
			tpl.ModelMParamDirName,
			tpl.ModelMRedisDirName,
		},
		[][]string{
			tpl.ModelApiMFilesName,
			tpl.ModelMApiFilesName,
			tpl.ModelMBaseFilesName,
			tpl.ModelMMongoFilesName,
			tpl.ModelMMysqlFilesName,
			tpl.ModelMParamFilesName,
			tpl.ModelMRedisFilesName,
		},
		[][]string{
			tpl.ModelApiMFilesContent,
			tpl.ModelMApiFilesContent,
			tpl.ModelMBaseFilesContent,
			tpl.ModelMMongoFilesContent,
			tpl.ModelMMysqlFilesContent,
			tpl.ModelMParamFilesContent,
			tpl.ModelMRedisFilesContent,
		},
	)
}

func initPConst(data ProjectData) {
	initModule(data, "pconst", []string{tpl.PConstDirName}, [][]string{tpl.PConstFilesName}, [][]string{tpl.PConstFilesContent})
}

func initGRpc(data ProjectData) {
	initModule(data, "grpc", []string{tpl.GRpcTestDirName}, [][]string{tpl.GRpcTestFilesName}, [][]string{tpl.GRpcTestFilesContent})
}

func initRoute(data ProjectData) {
	initModule(data, "route",
		[]string{
			tpl.RouteDirName,
			tpl.RouteV1DirName,
		},
		[][]string{
			tpl.RouteFilesName,
			tpl.RouteV1FilesName,
		},
		[][]string{
			tpl.RouteFilesContent,
			tpl.RouteV1FilesContent,
		},
	)
}

func initScript(data ProjectData) {
	initModule(data, "script",
		[]string{
			tpl.ScriptDirName,
			tpl.ScriptLogicDirName,
		},
		[][]string{
			tpl.ScriptFilesName,
			tpl.ScriptLogicFilesName,
		},
		[][]string{
			tpl.ScriptFilesContent,
			tpl.ScriptLogicFilesContent,
		},
	)
}

func initServer(data ProjectData) {
	initModule(data, "server",
		[]string{
			tpl.ServerDirName,
		},
		[][]string{
			tpl.ServerFilesName,
		},
		[][]string{
			tpl.ServerFilesContent,
		},
	)
}

func initService(data ProjectData) {
	initModule(data, "service", []string{tpl.ServiceDirName}, [][]string{tpl.ServiceFilesName}, [][]string{tpl.ServiceFilesContent})
}

func initMain(data ProjectData) {
	initModule(data, "main",
		[]string{
			tpl.MainDirName,
		},
		[][]string{
			tpl.MainFilesName,
		},
		[][]string{
			tpl.MainFilesContent,
		},
	)
}

func initProto(data ProjectData) {
	initModule(data, "proto",
		[]string{
			tpl.ProtoDirName,
		},
		[][]string{
			tpl.ProtoFilesName,
		},
		[][]string{
			tpl.ProtoFilesContent,
		},
	)
}
