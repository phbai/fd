package main

import (
	"fmt"
	"log"
	"os"

	"github.com/phbai/fd/drive"
	"github.com/urfave/cli/v2"
)

type FreeDrive struct {
	drive.Drive
}

func (fd *FreeDrive) Run() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "登录账户",
				Action: func(c *cli.Context) error {
					username := c.Args().First();
					password := c.Args().Get(1);
					err := fd.Login(username, password)
					if err != nil {
						fmt.Printf("%s,请重新输入\n", err)
					}

					return nil
				},
			},
			{
				Name:  "upload",
				Usage: "上传文件",
				Action: func(c *cli.Context) error {
					filename := c.Args().First()
					err := fd.Upload(filename)
					if err != nil {
						fmt.Printf("%s,请重新输入\n", err)
					}

					return nil
				},
			},
			{
				Name:  "download",
				Usage: "下载文件",
				Action: func(c *cli.Context) error {
					url := c.Args().First()
					err := fd.Download(url)
					if err != nil {
						fmt.Printf("%s,请重新输入\n", err)
					}

					return nil
				},
			},
			{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "查看信息",
				Action: func(c *cli.Context) error {
					url := c.Args().First()
					err := fd.Info(url)
					if err != nil {
						fmt.Printf("%s,请重新输入\n", err)
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
