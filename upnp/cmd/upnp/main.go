package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/MoZhonghua/mytools/upnp"
	"github.com/MoZhonghua/mytools/util"
	"github.com/codegangsta/cli"
)

var (
	devTypes = []string{
		"urn:schemas-upnp-org:device:InternetGatewayDevice:1",
		"urn:schemas-upnp-org:device:InternetGatewayDevice:2"}
)

func parseInt(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func marshalData(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "    ")
	return string(b)
}

var debug bool
var proxy string

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(-1)
}

func main() {
	app := cli.NewApp()
	app.Version = "1.0"
	app.Usage = "upnp client"
	app.Name = "upnp client"
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
			Usage:       "proxy",
			Destination: &proxy,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "listinterface",
			Usage:     "list all network interfaces",
			ArgsUsage: "",
			Action:    cmdListInterface,
		},
		{
			Name:      "ssdpsearch",
			Usage:     "SSDP Search",
			ArgsUsage: "<interface>",
			Action:    cmdSSDPSearch,
		},
		{
			Name:      "addportmap",
			Usage:     "add port mapping",
			ArgsUsage: "<igdURL> <localPort> <externalPort>",
			Action:    cmdAddPortMapping,
		},
		{
			Name:      "deleteportmap",
			Usage:     "delete port mapping",
			ArgsUsage: "<igdURL> <externalPort>",
			Action:    cmdDelPortMapping,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}

func discover(wg *sync.WaitGroup, intf *net.Interface, deviceType string,
	timeout time.Duration) error {
	defer wg.Done()
	results, err := upnp.SSDPSearch(intf, deviceType, timeout)
	if err != nil {
		fmt.Printf("failed to SSDP Search on %s: %v\n", intf.Name, err)
		return nil
	}
	for n := range results {
		b, _ := json.MarshalIndent(n, "", "    ")
		fmt.Println(string(b))
	}
	return nil
}

func createHttpClient() *util.HttpClient {
	cfg := &util.HttpClientConfig{}
	cfg.Proxy = proxy
	cfg.Debug = debug
	cfg.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	c, err := util.NewHttpClient(cfg)
	if err != nil {
		fail("error: %v", err)
	}
	return c
}

func cmdListInterface(c *cli.Context) error {
	intfs, err := net.Interfaces()
	if err != nil {
		fail("error: %v", err)
	}

	for _, intf := range intfs {
		fmt.Printf("%s %s\n", intf.Name, intf.Flags.String())
	}
	return nil
}

func cmdSSDPSearch(c *cli.Context) error {
	if len(c.Args()) < 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}

	intfName := c.Args()[0]

	intfs, err := net.Interfaces()
	if err != nil {
		fail("error: %v", err)
	}

	var wg sync.WaitGroup
	for _, intf := range intfs {
		if intfName != "all" && intfName != intf.Name {
			continue
		}
		fmt.Printf("intfName: %s %s\n", intf.Name, intf.Flags.String())
		if intf.Flags&net.FlagUp == 0 || intf.Flags&net.FlagMulticast == 0 {
			fmt.Printf("intfName down or multicast not support, skip\n")
			continue
		}

		for _, deviceType := range devTypes {
			wg.Add(1)
			discover(&wg, &intf, deviceType, 10*time.Second)
		}
	}
	wg.Wait()
	return nil
}

func cmdAddPortMapping(c *cli.Context) error {
	if len(c.Args()) < 3 {
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}

	hc := createHttpClient()

	igdURL := c.Args()[0]

	root, err := upnp.GetUPnPData(igdURL)
	if err != nil {
		fail("error: %v", err)
	}

	igd, err := upnp.GetIGDDevice(root, igdURL)
	if err != nil {
		fail("error: %v", err)
	}

	localPort, err := parseInt(c.Args()[1])
	if err != nil {
		fail("error: %v", err)
	}
	externalPort, err := parseInt(c.Args()[2])
	if err != nil {
		fail("error: %v", err)
	}

	for _, s := range igd.Services {
		externalIP, err := s.GetExternalIPAddress(hc)
		if err != nil {
			fail("error: %v", err)
		}

		err = s.AddPortMapping(hc, igd.LocalIPAddress.String(),
			"TCP", localPort, externalPort, "", 0)
		if err != nil {
			fail("error: %v", err)
		}

		fmt.Printf("Port mapping %s:%d -> %s:%d OK!\n",
			externalIP.String(), externalPort,
			igd.LocalIPAddress.String(), localPort)

		return nil
	}
	return nil
}

func cmdDelPortMapping(c *cli.Context) error {
	if len(c.Args()) < 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}

	igdURL := c.Args()[0]

	root, err := upnp.GetUPnPData(igdURL)
	if err != nil {
		fail("error: %v", err)
	}

	igd, err := upnp.GetIGDDevice(root, igdURL)
	if err != nil {
		fail("error: %v", err)
	}

	externalPort, err := parseInt(c.Args()[1])
	if err != nil {
		fail("error: %v", err)
	}

	hc := createHttpClient()
	for _, s := range igd.Services {
		err := s.DeletePortMapping(hc, "TCP", externalPort)
		if err != nil {
			fail("error: %v", err)
		}
	}
	fmt.Println("OK!")
	return nil
}
