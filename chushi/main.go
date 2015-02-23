package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {

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
			},
		},
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "create something",
			Action: func(c *cli.Context) {
				println("added task: ", c.Args().First())
			},
			Subcommands: []cli.Command{
				{
					Name:      "kitchen",
					ShortName: "k",
					Usage:     "create a kitchen",
					Action: func(c *cli.Context) {
						println("new task template: ", c.Args().First())
					},
				},
				{
					Name:      "router",
					ShortName: "r",
					Usage:     "create a router",
					Action: func(c *cli.Context) {
						println("removed task template: ", c.Args().First())
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
