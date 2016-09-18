package template

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

var debug bool
var proxy string

func marshalData(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

func exitOnError(err error) {
	if err == nil {
		return
	}
	fail("error: %v", err)
}

func showHelp(c *cli.Context) {
	cli.ShowSubcommandHelp(c)
	os.Exit(1)
}

func main() {
	app := cli.NewApp()
	app.Version = "1.0"
	app.Usage = ""
	app.Name = ""
	app.Author = "MoZhonghua"
	app.CommandNotFound = func(ctx *cli.Context, command string) {
		fail("unknown command: %v", command)
	}
	app.Writer = os.Stdout

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug",
			Destination: &debug,
		},
		&cli.StringFlag{
			Name:        "proxy",
			Usage:       "proxy address",
			Value:       "",
			Destination: &proxy,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "echo",
			Usage:     "echo users input",
			ArgsUsage: "[content]",
			Action:    cmdEcho,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func cmdEcho(c *cli.Context) error {
	if len(c.Args()) == 0 {
		showHelp(c)
	}
	return nil
}
