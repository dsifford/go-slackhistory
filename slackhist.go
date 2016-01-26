package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type message struct {
	User, Type, Subtype, Text, Ts string
	Timestamp                     time.Time
}

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
			processData(c.String("single"))
		}
	}

	app.Run(os.Args)
}

func processData(f string) map[string][]message {

	payload := make(map[string][]message)
	var dirname string

	// Open a readable stream from teh zip file
	r, err := zip.OpenReader(f)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// For every item in the entire zip folder (recursive)
	for _, file := range r.File {

		isDirectory := file.FileInfo().IsDir()

		// Switch on file = directory
		switch isDirectory {
		case true:

			// Retreive the current directory's base name
			path, err := filepath.Abs(file.FileHeader.Name)
			if err != nil {
				panic(err)
			}
			dirname = filepath.Base(path)

			// Make the directory's base name (thus, the channel name) the key
			// and instantiate an empty slice within payload for the messages to
			// be stored
			payload[dirname] = make([]message, 0)

		case false:
			var thisJSON []message

			// Grab the parent directory's name
			dirname = filepath.Base(filepath.Dir(file.FileHeader.Name))

			// Open a readable stream from the current file
			rc, err := file.Open()
			if err != nil {
				panic(err)
			}
			defer rc.Close()

			// Create JSON decoder and decode the readable stream to "thisJSON"
			decoder := json.NewDecoder(rc)
			for decoder.More() {
				if err := decoder.Decode(&thisJSON); err != nil {
					panic(err)
				}
			}

			// After a full file decode, iterate through the decoded JSON for
			// messages or attachments, and parse the timezone according to the
			// current user's local time zone format
			for _, value := range thisJSON {
				if value.Type == "message" && value.Subtype == "" || value.Subtype == "file_share" {

					// TODO: Convert User from user ID string to actual name
					// TODO: In messages containing mentions, replace the mentioned
					// user's raw ID with their name

					// Enforce strict checking on target var
					var timestamp time.Time

					timestring, err := strconv.ParseInt(strings.Split(value.Ts, ".")[0], 10, 64)
					if err != nil {
						panic(err)
					}
					tm := time.Unix(timestring, 0)

					// TODO: Eventually make a flag that allows the user to override
					// the default time zone to whatever time zone they choose
					loc, err := time.LoadLocation("Local")
					if err != nil {
						panic(err)
					}

					timestamp = tm.In(loc)
					value.Timestamp = timestamp

					payload[dirname] = append(payload[dirname], value)
				}
			}

		}

	}

	// NOTE: This iteration is for debugging purposes only -- remove from final
	for masterkey, mastervalue := range payload {
		fmt.Printf("======================KEY======================\n%v\n===============================================\n\n\n", masterkey)
		for key, value := range mastervalue {
			fmt.Printf("Key: %v\n\n"+
				"Timestamp: %v\nUser: %v\nMessage: %v\n----------\n\n", key, value.Timestamp, value.User, value.Text)
		}
	}

	return payload

}
