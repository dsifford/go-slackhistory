package main

import (
	"archive/zip"
	// "bufio"
	// "encoding/json"
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
			unzip(c.String("single"))
		}
	}

	app.Run(os.Args)
}

func unzip(f string) {
	r, err := zip.OpenReader(f)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	for _, f := range r.File {

		if f.FileInfo().IsDir() {
			// Add flow control here to extract all files within
			// a specific dir (channel) to its own CSV
			// (make this into another function)
			fmt.Println(f.Name)
		}

	}

}

// type Message struct {
// 	User, Type, Subtype, Text, Ts string
// }
// decoder := json.NewDecoder(rc)
//
// var m Message
// for decoder.More() {
// 	err = decoder.Decode(&m)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	fmt.Printf("%v: %v\n", m.User, m.Text)
// }

// rc, err := f.Open()
// if err != nil {
// 	panic(err)
// }
// scanner := bufio.NewScanner(rc)
// for scanner.Scan() {
// 	fmt.Println(scanner.)
// }
