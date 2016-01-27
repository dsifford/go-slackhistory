package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	// "github.com/codegangsta/cli"
	// "os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type message struct {
	User, Type, Subtype, Text, Ts string
	Timestamp                     time.Time
}

type meta struct {
	ID, Name string
	RealName string `json:"real_name,omitempty"`
}

// TODO: Improve overall concurrency / reliability

func main() {
	metadata := make(map[string][]meta)
	processData("temp/export.zip", metadata)

	// NOTE: Turning this off for now until the main function of this application
	// is working and ready to add additional flags/features
	//
	// app := cli.NewApp()
	// app.Usage = "An app for exporting Slack history to CSV"
	//
	// app.Flags = []cli.Flag{
	// 	cli.StringFlag{
	// 		Name: "single, s",
	// 	},
	// }
	//
	// app.Action = func(c *cli.Context) {
	// 	if c.String("single") != "" {
	// 		fmt.Println("Exporting a single channel...")
	// 		processData(c.String("single"), metadata)
	// 	}
	// }
	//
	// app.Run(os.Args)
}

// FIXME: Concurrency issue with this function. Add data channels/workers to make
// sure that the process isn't ending before the worker is done.
func processData(f string, md map[string][]meta) map[string][]message {

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

			// Switch on filename
			switch file.FileHeader.Name {
			case "integration_logs.json":
				break
			case "channels.json", "users.json":
				getMeta(file, md)
				break
			}

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

					value.Timestamp = parseTimestamp(value.Ts)
					value.User = parseUser(value.User, md)

					re := regexp.MustCompile("(<@[a-zA-Z0-9]{9}(\\p{S}[a-zA-Z0-9]+)?>)")
					if matches := re.FindAllString(value.Text, -1); matches != nil {

						value.Text = re.ReplaceAllStringFunc(value.Text, func(match string) string {
							uid := match[2 : len(match)-1]
							return "@" + parseUserShortName(uid, md)
						})
						fmt.Println(value.Text)

					}

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

func getMeta(f *zip.File, m map[string][]meta) {
	filename := f.FileHeader.Name

	var thisJSON []meta

	rc, err := f.Open()
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	decoder := json.NewDecoder(rc)
	for decoder.More() {
		if err := decoder.Decode(&thisJSON); err != nil {
			panic(err)
		}
	}

	switch filename {
	case "users.json":
		m["users"] = thisJSON
	case "channels.json":
		m["channels"] = thisJSON
	default:
		panic(filename)
	}

}

func parseTimestamp(ts string) time.Time {

	timestring, err := strconv.ParseInt(strings.Split(ts, ".")[0], 10, 64)
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

	return tm.In(loc)
}

func parseUser(uid string, md map[string][]meta) string {
	var username string
	for _, value := range md["users"] {
		if value.ID == uid {
			username = value.RealName
			break
		}
	}
	return username
}

// TODO: Delete this function and combine it with above
func parseUserShortName(uid string, md map[string][]meta) string {
	var username string
	for _, value := range md["users"] {
		if value.ID == uid {
			username = value.Name
			break
		}
	}
	return username
}
