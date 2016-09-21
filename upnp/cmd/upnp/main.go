package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MoZhonghua/mytools/upnp"
	"github.com/MoZhonghua/mytools/util"
	"gopkg.in/urfave/cli.v1"
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
			Name:      "ssdp",
			Usage:     "SSDP Search",
			ArgsUsage: "<interface> [destIP]",
			Action:    cmdSSDPSearch,
		},
		{
			Name:      "addportmap",
			Usage:     "add port mapping",
			ArgsUsage: "<igdDeviceIP> <localIP:localPort> <externalPort>",
			Action:    cmdAddPortMapping,
		},
		{
			Name:      "deleteportmap",
			Usage:     "delete port mapping",
			ArgsUsage: "<igdDeviceIP> <externalPort>",
			Action:    cmdDelPortMapping,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
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

func doSearch(wg *sync.WaitGroup, intf *net.Interface, deviceType string,
	ip net.IP, timeout time.Duration) error {
	defer wg.Done()
	results, err := upnp.SSDPSearch(intf, deviceType, ip, timeout)
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

func cmdSSDPSearch(c *cli.Context) error {
	if len(c.Args()) < 1 {
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}

	intfName := c.Args()[0]
	destIP := upnp.SSDPMulticastAddr
	if len(c.Args()) >= 2 {
		destIP = net.ParseIP(c.Args()[1])
	}

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
			doSearch(&wg, &intf, deviceType, destIP, 5*time.Second)
		}
	}
	wg.Wait()
	return nil
}

func parseIPPort(s string) (net.IP, int, error) {
	f := strings.Split(s, ":")
	if len(f) != 2 {
		return nil, 0, fmt.Errorf("invalid addr: %s", s)
	}

	ip := net.ParseIP(f[0])
	port, err := parseInt(f[1])
	if err != nil {
		return nil, 0, err
	}

	return ip, port, nil
}

func cmdAddPortMapping(c *cli.Context) error {
	if len(c.Args()) < 3 {
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}

	hc := createHttpClient()

	igdDeviceIP := c.Args()[0]
	igdURL := fmt.Sprintf("http://%s:1900/igd.xml", igdDeviceIP)

	root, err := upnp.GetUPnPData(igdURL)
	if err != nil {
		fail("error: %v", err)
	}

	igd, err := upnp.GetIGDDevice(root, igdURL)
	if err != nil {
		fail("error: %v", err)
	}

	localIP, localPort, err := parseIPPort(c.Args()[1])
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

		err = s.AddPortMapping(hc, localIP.String(),
			"TCP", localPort, externalPort, "", 0)
		if err != nil {
			fail("error: %v", err)
		}

		fmt.Printf("Port mapping %s:%d -> %s:%d OK!\n",
			externalIP.String(), externalPort,
			localIP.String(), localPort)

		return nil
	}
	return nil
}

func cmdDelPortMapping(c *cli.Context) error {
	if len(c.Args()) < 2 {
		cli.ShowSubcommandHelp(c)
		os.Exit(1)
	}

	igdDeviceIP := c.Args()[0]
	igdURL := fmt.Sprintf("http://%s:1900/igd.xml", igdDeviceIP)

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
