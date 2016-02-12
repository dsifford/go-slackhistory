[![Go Report Card](https://goreportcard.com/badge/github.com/dsifford/slackhist)](https://goreportcard.com/report/github.com/dsifford/slackhist)
## slackhist
A command-line utility for exporting Slack history to Excel (`.xlsx`)

### Installation
**Have Go installed on your system?**
```sh
go get github.com/dsifford/slackhist
cd $GOPATH/src/github.com/dsifford
go install
```
**Don't have Go installed?**
- **OSX/Linux**: Download a pre-built binary from [releases](https://github.com/dsifford/slackhist/releases) and save the file to your `usr/local/bin` directory.
- **Windows**: Download a pre-built executable file from [releases](https://github.com/dsifford/slackhist/releases), save it wherever you'd like, and include that directory in your `PATH`.

### Usage
![helptext](http://i.imgur.com/xSlguN5.png)

#### Basic Usage
1. Save the exported `.zip` archive from Slack to a memorable location (eg, `~/Downloads/export.zip`).
2. Open a terminal and navigate to the `.zip` file location (eg, `cd ~/Downloads`).
3. Enter `slackhist <YOUR-ZIP-FILE-NAME>` (eg, `slackhist export.zip`).
4. The new `.xlsx` file will now be in your current directory.

#### CLI Global Options
- `-n, --name`: Renames the output file (Default `YYYY-MMM-DD_SlackExport.xlsx`)
- `-d, --destination`: Changes the output directory (Default: the current working directory)
- `-t, --timezone`: Changes the time-zone parsing of each message timestamp (Default: your local timezone)

### Benchmarks (34 Channels, 824 Messages)
#### Version 0.0.0
|real|user|sys|
|---|---|---|
`0m0.846s`|`0m1.128s`|`0m0.088s`

### Todo...
- [ ] Refactor
- [ ] Improve concurrency
- [ ] Add tests
- [ ] Compile for Windows (64 and 32 bit)
