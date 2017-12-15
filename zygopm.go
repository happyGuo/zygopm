package main

import (
	"os"

	"github.com/urfave/cli"

	"zygopm/cmd"
	"zygopm/module/conf"
	"zygopm/module/msg"
	"zygopm/module/repo"
	"zygopm/module/setting"
	"path/filepath"
)

var version = "0.1.0-dev"

func main() {
	app := cli.NewApp()
	app.Usage = `克隆到$GOPATH/src目录下 go build zygopm.go 生成可执行文件`
	app.Name = "zygopm"
	app.Version = version
	//cli 执行的命令
	app.Commands = commands()

	app.Run(os.Args)
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:  "init",
			Usage: "初始化一个新项目，创建依赖配置文件",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "c",
					Usage: "指定配置文件name 不用加后缀",
				},
			},

			Action: func(c *cli.Context) error {
				if c.String("c")!="" {
					setting.SetGopmConfFIleName(c.String("c"))
				}
				cmd.Create(".")
				return nil
			},
		},
		{
			Name:  "i",
			Usage: "安装依赖包到指定目录，不指定则在GOAPTH的第一个目录",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "p",
					Usage: "安装到指定目录",
				},

				cli.BoolFlag{
					Name:  "v",
					Usage: "更新到项目vendor目录",
				},

				cli.StringFlag{
					Name:  "c",
					Usage: "指定配置文件name 不用加后缀",
				},
			},

			Action: func(c *cli.Context) error {

				if c.Bool("p") && c.Bool("v") {
					msg.Die("不能同时指定 -p -v")
				}
				iPath := setting.GetFirstGOPATH()
				if c.Bool("p") {
					iPath = conf.GetInstallPath()
					msg.Warn("依赖包将会安装到=>" + iPath)
				}

				if c.Bool("v") {
					iPath = filepath.Join(".", setting.VENDOR)
					msg.Warn("依赖包将会安装到=>" + iPath)
				}

				if c.String("c")!="" {
					setting.SetGopmConfFIleName(c.String("c"))
					setting.SetGopmLockFIleName(c.String("c"))
				}

				installer := repo.NewInstaller()
				//设置安装位置
				installer.Vendor = iPath
				cmd.Install(installer)
				return nil
			},
		},

		{
			Name:  "up",
			Usage: "更新依赖包",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "p",
					Usage: "更新指定目录的包",
				},
				cli.BoolFlag{
					Name:  "v",
					Usage: "更新到项目vendor目录",
				},
				cli.StringFlag{
					Name:  "c",
					Usage: "指定配置文件name 不用加后缀",
				},
			},

			Action: func(c *cli.Context) error {

				if c.Bool("p") && c.Bool("v") {
					msg.Die("不能同时指定 -p -v")
				}
				iPath := setting.GetFirstGOPATH()
				if c.Bool("p") {
					iPath = conf.GetInstallPath()
					msg.Warn("依赖包将会安装到=>" + iPath)
				}

				if c.Bool("v") {
					iPath = filepath.Join(".", setting.VENDOR)
					msg.Warn("依赖包将会安装到=>" + iPath)
				}

				if c.String("c")!="" {
					setting.SetGopmConfFIleName(c.String("c"))
					setting.SetGopmLockFIleName(c.String("c"))
				}

				installer := repo.NewInstaller()
				//设置安装位置
				installer.Vendor = iPath
				cmd.Update(installer)
				return nil
			},
		},
	}
}



