package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/MoZhonghua/mytools/tcpmux"
	"github.com/codegangsta/cli"
)

var debug bool
var proxy string
var server string

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

var logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

func main() {
	app := cli.NewApp()
	app.Version = "1.0"
	app.Usage = ""
	app.Name = "tcpmux-admin tool"
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
		&cli.StringFlag{
			Name:        "server",
			Usage:       "server admin address",
			Value:       "http://127.0.0.1:6732",
			Destination: &server,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "list",
			Usage:  "list target",
			Action: cmdList,
		},
		{
			Name:      "add",
			Usage:     "add target",
			ArgsUsage: "<id> <target(ip:port)>",
			Action:    cmdAdd,
		},
		{
			Name:      "batch",
			Usage:     "add all targets defined in a json file",
			ArgsUsage: "<file>",
			Action:    cmdBatch,
		},
		{
			Name:      "delete",
			Usage:     "delete target",
			ArgsUsage: "<id>",
			Action:    cmdDelete,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func createClient() *tcpmux.Client {
	client, err := tcpmux.NewClient(server, logger, proxy, debug)
	exitOnError(err)
	return client
}

func cmdAdd(c *cli.Context) error {
	if len(c.Args()) < 2 {
		showHelp(c)
	}
	id := c.Args()[0]
	target := c.Args()[1]

	client := createClient()
	err := client.AddTarget(id, target)
	exitOnError(err)

	fmt.Println("OK!")
	return nil
}

func cmdBatch(c *cli.Context) error {
	if len(c.Args()) < 1 {
		showHelp(c)
	}

	path := c.Args()[0]
	data, err := ioutil.ReadFile(path)
	exitOnError(err)

	list := make([]*tcpmux.TargetInfo, 0)
	err = json.Unmarshal(data, &list)
	exitOnError(err)

	client := createClient()
	for _, pm := range list {
		err := client.AddTarget(pm.Id, pm.Target)
		if err != nil {
			fmt.Printf("%6s -> %s: %v\n", pm.Id, pm.Target, err)
		} else {
			fmt.Printf("%6s -> %s: OK!\n", pm.Id, pm.Target)
		}
	}

	return nil
}

func cmdDelete(c *cli.Context) error {
	if len(c.Args()) < 1 {
		showHelp(c)
	}
	id := c.Args()[0]

	client := createClient()
	err := client.DeleteTarget(id)
	exitOnError(err)

	fmt.Println("OK!")
	return nil
}

func cmdList(c *cli.Context) error {
	client := createClient()
	m, err := client.ListTarget()
	exitOnError(err)

	fmt.Println(marshalData(m))
	return nil
}
