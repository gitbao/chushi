package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/gitbao/chushi/ec2"
	"github.com/gitbao/chushi/model"
	"github.com/gitbao/chushi/shell"
)

func main() {

	err := model.DB.DB().Ping()
	if err != nil {
		fmt.Println("Error pinging database. Do you have a postgres database set up called chushi?")
		return
	}
	app := cli.NewApp()
	app.Name = "Chushi"
	app.Usage = "The bao chef. Configures and manages the kitchen, routers, and xialongbao"
	app.Action = func(c *cli.Context) {
		println("Ready to make some tasty buns?")
	}

	app.Commands = []cli.Command{
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "List everything",
			Action: func(c *cli.Context) {
				println("added task: ", c.Args().First())
				ec2.CreateInstance()
			},
		},
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "create something",
			Action: func(c *cli.Context) {
				kind := c.Args().First()
				if kind != "kitchen" && kind != "router" && kind != "xialong" {
					fmt.Println("You need to specify a server type.",
						"Choose kitchen, router, or xiaolong.")
					return
				}
				fmt.Println("Creating a", kind, "server:")
				fmt.Println("Creating instance:")
				// server := ec2.CreateInstance()
				server := model.Server{
					Ip: "54.146.69.23",
				}
				shell.RunShellScript(kind, &server)

			},
		},
	}

	app.Run(os.Args)
}
