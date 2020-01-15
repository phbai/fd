package main

import (
	"fmt"
	"log"
	"os"

	"github.com/phbai/FreeDrive/drive"
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
					fmt.Println("added task: ", c.Args().First())
					return nil
				},
			},
			{
				Name:  "upload",
				Usage: "上传文件",
				Action: func(c *cli.Context) error {
					fmt.Println("completed task: ", c.Args().First())
					return nil
				},
			},
			{
				Name:  "download",
				Usage: "下载文件",
				Action: func(c *cli.Context) error {
					fmt.Println("url: ", c.Args().First())
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