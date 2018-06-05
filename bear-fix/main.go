package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type Command struct {
	Arguments []string `json:"arguments"`
	Directory string   `json:"directory"`
	File      string   `json:"file"`
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("usage: bear-fix <compile_commands_file>")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	commands := make([]*Command, 0)
	err = json.Unmarshal(data, &commands)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, c := range commands {
		c.Arguments = remove(c.Arguments, c.File)
		c.Arguments = remove(c.Arguments, "cc")
		c.Arguments = remove(c.Arguments, "clang")
		c.Arguments = remove(c.Arguments, "c++")
	}

	data, err = json.MarshalIndent(commands, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(flag.Arg(0), data, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
