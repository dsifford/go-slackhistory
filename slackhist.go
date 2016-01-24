package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Usage = "An app for exporting Slack history to CSV"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "single, s",
		},
	}

	app.Action = func(c *cli.Context) {
		if c.String("single") != "" {
			fmt.Println("Exporting a single channel...")
		}
	}

	app.Run(os.Args)
}
