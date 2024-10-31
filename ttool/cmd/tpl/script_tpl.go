package tpl

var (
	ScriptDirName   = "script"
	ScriptFilesName = []string{
		"commands.go",
	}
	ScriptFilesContent = []string{
		ScriptCommandsGo,
	}

	ScriptLogicDirName   = "script/logic"
	ScriptLogicFilesName = []string{
		"init.go",
		"tests.go",
	}
	ScriptLogicFilesContent = []string{
		ScriptLogicInitGo,
		ScriptLogicTestsGo,
	}
)

var (
	ScriptCommandsGo = `package script

import (
	"{{.Name}}/script/logic"

	"github.com/urfave/cli"
)

func Commands() []cli.Command {
	return []cli.Command{
		// test application
		{
			Name:  "test_daemon",
			Usage: "daemon 这是一个测试命令，供测试使用 2016-12-12",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "lang",
					Value: "english",
					Usage: "language for the greeting",
				},
			},
			Action: func(c *cli.Context) error {
				ps := ""
				if c.NArg() > 0 {
					ps = c.Args()[0]
				} else {
					return cli.NewExitError("Test script error.", 1)
				}
				lang := c.String("lang")
				logic.TestDaemon(lang, ps)
				println("Test script over")
				return nil
			},
		},

		// test application
		{
			Name:  "test_script",
			Usage: "script 这是一个测试命令，供测试使用 2016-12-12",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "lang",
					Value: "english",
					Usage: "language for the greeting",
				},
			},
			Action: func(c *cli.Context) error {
				ps := ""
				if c.NArg() > 0 {
					ps = c.Args()[0]
				} else {
					return cli.NewExitError("Test script error.", 1)
				}
				lang := c.String("lang")
				logic.TestScript(lang, ps)
				println("Test script over")
				return nil
			},
		},
	} // end []cli.Command
} // end Commands()
`
	ScriptLogicInitGo = `package logic

import (
	"fmt"
	"github.com/namejlt/gozen"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	ServerType     string
	workStatusFlag bool
)

func init() {
	//环境初始化

	ServerType = gozen.ConfigAppGetString("ServerType", "script")
	workStatusFlag = true
	if ServerType == "script" {
		go initWorkStatusFlag()
	}
}

func initWorkStatusFlag() {
	var sigs []os.Signal
	sigs = append(sigs, syscall.SIGINT)  //2
	sigs = append(sigs, syscall.SIGQUIT) //3
	sigs = append(sigs, syscall.SIGKILL) //9
	sigs = append(sigs, syscall.SIGTERM) //15

	c := make(chan os.Signal)
	signal.Notify(c, sigs...)
	sig := <-c
	signal.Stop(c)
	close(c)

	fmt.Println("handle signal: ", sig)
	workStatusFlag = false
	fmt.Println("script workStatusFlag false, time:", time.Now().String())
}
`
	ScriptLogicTestsGo = `package logic

import (
	"fmt"
	"time"
)

func TestDaemon(lang, ps string) {
	for workStatusFlag {
		hi := "hello"
		switch lang {
		case "chinese":
			hi = "你好"
		case "english":
			hi = "hello"
		}
		fmt.Println(hi + " : " + ps)
		fmt.Println(time.Now().UnixNano())
		time.Sleep(time.Duration(5) * time.Second)
	}
}

func TestScript(lang, ps string) {
	hi := "hello"
	switch lang {
	case "chinese":
		hi = "你好"
	case "english":
		hi = "hello"
	}
	fmt.Println(hi + " : " + ps)
	fmt.Println(time.Now().UnixNano())
}
`
)
