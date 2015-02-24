package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	tm "github.com/buger/goterm"
	"github.com/codegangsta/cli"
	"github.com/gitbao/chushi/ec2"
	"github.com/gitbao/chushi/shell"
	"github.com/gitbao/gitbao/model"
)

const (
	isProduction = false
)

func getServer(args) (server model.Server) {
	serverId, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal(err)
		return
	}
	query := model.DB.Find(&server, int64(serverId))
	if query.Error != nil {
		log.Fatal(query.Error)
		return
	}
	return server
}

func main() {
	err := model.DB.DB().Ping()
	if err != nil {
		fmt.Println("Error pinging database. Do you have a postgres database set up called chushi?")
		return
	}
	if isProduction {
		os.Setenv("GO_ENV", "production")
		model.Close()
		model.Connect()
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
				var servers []model.Server
				query := model.DB.Find(&servers)
				if query.Error != nil {
					log.Fatal(query.Error)
				}

				totals := tm.NewTable(0, 10, 5, ' ', 0)
				fmt.Fprintf(totals, "Id\tInstanceId\tKind\tIp\n")

				for _, value := range servers {
					fmt.Fprintf(totals, "%d\t%s\t%s\t%s\n", value.Id, value.InstanceId, value.Kind, value.Ip)
				}
				tm.Println(totals)
				tm.Flush()

			},
		},
		{
			Name:      "new",
			ShortName: "n",
			Usage:     "spin up a new ec2 server",
			Action: func(c *cli.Context) {
				fmt.Println("Creating instance:")
				server := ec2.CreateInstance()
				model.DB.Create(&server)
				fmt.Println("Created server with Id:", server.Id)
			},
		},
		{
			Name:      "assign",
			ShortName: "a",
			Usage:     "assign a server type to a server",
			Action: func(c *cli.Context) {
				args := c.Args()
				server := getServer(args)

				kind := args[1]
				fmt.Println("Assigning a", kind, "server:")
				fmt.Println("Created server with Id:", server.Id)
				shell.Initialize(kind, &server)
				server.Kind = kind
				model.DB.Save(&server)
			},
		},
		{
			Name:      "update",
			ShortName: "u",
			Usage:     "update a server",
			Action: func(c *cli.Context) {
				args := c.Args()

				server := getServer(args)

				fmt.Println("Created server with Id:", server.Id)
				shell.Update(server.Kind, &server)
			},
		},
		{
			Name:  "open",
			Usage: "create something",
			Action: func(c *cli.Context) {
				args := c.Args()

				server := getServer(args)

				cmd := exec.Command("open", "http://"+server.Ip+":8000")
				err = cmd.Run()
				if err != nil {
					log.Fatal(err)
					return
				}

			},
		},
		{
			Name:  "logs",
			Usage: "",
			Action: func(c *cli.Context) {
				args := c.Args()

				server := getServer(args)

				fmt.Println("Logs for server:", server.Id)
				shell.Logs(&server)
			},
		},
		{
			Name:  "destroy",
			Usage: "destroy a server",
			Action: func(c *cli.Context) {
				args := c.Args()

				server := getServer(args)

				err = ec2.DestroyInstance(server.InstanceId)
				if err != nil {
					log.Fatal(err)
					return
				}
				model.DB.Delete(&server)
			},
		},
	}

	app.Run(os.Args)
}
