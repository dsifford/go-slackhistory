package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/tealeg/xlsx"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type message struct {
	User, Type, Subtype, Text, Ts string
	Timestamp                     time.Time
}

// Implement the sort interface for type messages
type messages []message

func (m messages) Len() int {
	return len(m)
}

func (m messages) Less(i, j int) bool {
	return m[i].Timestamp.Before(m[j].Timestamp)
}

func (m messages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type meta struct {
	ID, Name string
	RealName string `json:"real_name,omitempty"`
}

// Set global CLI vars
var filename, outputDir, timezone string
var metadata = make(map[string][]meta)

func main() {

	app := cli.NewApp()
	app.Usage = "A command-line utility for exporting Slack history to Excel (.xlsx)"
	app.HideVersion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "name, n",
			Value:       time.Now().Format("2006-Jan-02") + "_SlackExport.xlsx",
			Usage:       "Set the name of the exported spreadsheet",
			Destination: &filename,
		},
		cli.StringFlag{
			Name:        "destination, d",
			Value:       "./",
			Usage:       "Specify the output directory for\n\t  the exported xlsx workbook",
			Destination: &outputDir,
		},
		cli.StringFlag{
			Name:        "timezone, t",
			Value:       "Local",
			Usage:       "Specify an alternate timezone.\n\t  See here for available options https://goo.gl/Tmq0oR",
			Destination: &timezone,
		},
	}

	app.Action = func(c *cli.Context) {
		if len(c.Args()) != 1 {
			fmt.Printf("\n*** Error: The location of the target zip file must be the last and only argument!\n***\n" +
				"*** For additional help, enter \"slackhist help\"\n\n")
			os.Exit(1)
		}

		if suffix := strings.HasSuffix(filename, ".xlsx"); suffix == false {
			filename = filename + ".xlsx"
		}

		messages := processData(c.Args()[0], metadata)
		createWorkbook(messages)
	}

	app.Run(os.Args)
}

func processData(f string, md map[string][]meta) map[string]messages {

	payload := make(map[string]messages)
	var dirname string

	// Open a readable stream from the zip file
	r, err := zip.OpenReader(f)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// Collect the required meta to process the information HACK
	for _, file := range r.File {
		switch file.FileHeader.Name {
		case "channels.json", "users.json":
			getMeta(file, md)
		}
	}

	// For every item in the entire zip folder (recursive)
	for _, file := range r.File {

		isDirectory := file.FileInfo().IsDir()

		// Switch on file = directory
		switch isDirectory {
		case true:

			// Retrieve the current directory's base name
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

			// If the file is one that we've already processed, break
			// from this particular loop cycle
			switch file.FileHeader.Name {
			case "integration_logs.json", "channels.json", "users.json":
				continue
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
					_, value.User = parseUser(value.User, md)

					re := regexp.MustCompile(`(<@[a-zA-Z0-9]{9}(\|[a-zA-Z0-9._]+)?>)`)
					if matches := re.FindAllString(value.Text, -1); matches != nil {

						value.Text = re.ReplaceAllStringFunc(value.Text, func(match string) string {
							uid := match[2 : len(match)-1]
							if value.Subtype == "file_share" {
								uid = strings.Split(uid, "|")[0]
							}
							username, _ := parseUser(uid, md)
							return "@" + username
						})

					}

					payload[dirname] = append(payload[dirname], value)
				}
			}

		}

	}

	return payload

}

func createWorkbook(m map[string]messages) {

	var channelNames sort.StringSlice

	workbook := xlsx.NewFile()
	fullPath := filepath.Clean(filepath.Join(outputDir, filename))

	// First, sort the map by channel name
	for channel := range m {
		channelNames = append(channelNames, channel)
	}

	channelNames.Sort()

	for _, channel := range channelNames {
		sheet, err := workbook.AddSheet(channel)
		if err != nil {
			panic(err)
		}

		sort.Sort(sort.Reverse(m[channel]))

		headingRow := sheet.AddRow()
		headingRow.AddCell().SetString("Timestamp")
		headingRow.AddCell().SetString("User")
		headingRow.AddCell().SetString("Message")

		if err := sheet.SetColWidth(0, 1, 20); err != nil {
			panic(err)
		}

		if err := sheet.SetColWidth(2, 2, 200); err != nil {
			panic(err)
		}

		for _, message := range m[channel] {
			messageRow := sheet.AddRow()
			messageRow.AddCell().SetString(message.Timestamp.Format("Jan 02, 2006 | 15:04"))
			messageRow.AddCell().SetString(message.User)
			messageRow.AddCell().SetString(message.Text)
		}

	}

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		panic(err)
	}

	if err := workbook.Save(fullPath); err != nil {
		panic(err)
	}

}

/**************************************************************************
 *                         Utility functions                              *
 **************************************************************************/

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

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	return tm.In(loc)
}

func parseUser(uid string, m map[string][]meta) (string, string) {
	var username, realname string
	for _, value := range m["users"] {
		if value.ID == uid {
			username = value.Name
			realname = value.RealName
			break
		}
	}
	return username, realname
}
