package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/MoZhonghua/mytools/tcpproxy"
	"github.com/MoZhonghua/mytools/util"
	"gopkg.in/urfave/cli.v1"
)

var debug bool
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
			Name:        "server",
			Usage:       "server admin address",
			Value:       "http://127.0.0.1:3333",
			Destination: &server,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "list",
			Usage:  "list port mapping",
			Action: cmdList,
		},
		{
			Name:   "dump",
			Usage:  "dump port mapping",
			Action: cmdDump,
		},
		{
			Name:      "add",
			Usage:     "add port mapping",
			ArgsUsage: "<localPort> <remoteAddr(ip:port)>",
			Action:    cmdAdd,
		},
		{
			Name:      "delete",
			Usage:     "delete port mapping",
			ArgsUsage: "<localPort>",
			Action:    cmdDelete,
		},
		{
			Name:      "batch",
			Usage:     "add all port mappings defined in a json file",
			ArgsUsage: "<file>",
			Action:    cmdBatch,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func createClient() *tcpproxy.Client {
	util.SetHttpClientDebugMode(debug)
	client, err := tcpproxy.NewClient(server)
	exitOnError(err)
	return client
}

func cmdAdd(c *cli.Context) error {
	if len(c.Args()) < 2 {
		showHelp(c)
	}
	localPort, err := strconv.ParseInt(c.Args()[0], 10, 32)
	exitOnError(err)
	remoteAddr := c.Args()[1]

	client := createClient()
	err = client.AddPortMapping(int(localPort), remoteAddr)
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

	list := make([]*tcpproxy.PortMappingInfo, 0)
	err = json.Unmarshal(data, &list)
	exitOnError(err)

	client := createClient()
	for _, pm := range list {
		err := client.AddPortMapping(pm.LocalPort, pm.RemoteAddr)
		if err != nil {
			fmt.Printf("%5d -> %s: %v\n", pm.LocalPort, pm.RemoteAddr, err)
		} else {
			fmt.Printf("%5d -> %s: OK!\n", pm.LocalPort, pm.RemoteAddr)
		}
	}

	return nil
}

func cmdDelete(c *cli.Context) error {
	if len(c.Args()) < 1 {
		showHelp(c)
	}
	localPort, err := strconv.ParseInt(c.Args()[0], 10, 32)
	exitOnError(err)

	client := createClient()
	err = client.DeletePortMapping(int(localPort))
	exitOnError(err)

	fmt.Println("OK!")
	return nil
}

func cmdList(c *cli.Context) error {
	client := createClient()
	m, err := client.ListPortMapping()
	exitOnError(err)

	for _, p := range m {
		fmt.Printf("%-5d -> %s\n", p.LocalPort, p.RemoteAddr)
	}

	return nil
}

func cmdDump(c *cli.Context) error {
	client := createClient()
	m, err := client.ListPortMapping()
	exitOnError(err)

	fmt.Println(marshalData(m))
	return nil
}
