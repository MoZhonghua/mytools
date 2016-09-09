package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type compileFlag struct {
	flag string
	arg  string
}

func isSourceFile(s string) bool {
	if strings.HasSuffix(s, ".c") ||
		strings.HasSuffix(s, ".cpp") ||
		strings.HasSuffix(s, ".c++") ||
		strings.HasSuffix(s, ".o") {
		return true
	}
	return false
}

func GenereteYcmFlags(cmd string) (string, error) {
	gccIndex := strings.Index(cmd, "gcc")
	if gccIndex == -1 {
		return "", errors.New("no gcc or g++")
	}

	cmd = cmd[gccIndex:]
	semiColonIndex := strings.Index(cmd, ";")
	if semiColonIndex != -1 {
		cmd = cmd[:semiColonIndex]
	}

	fmt.Println(cmd)
	parts := strings.Split(cmd, " ")
	fmt.Printf("flags = [\n")
	skip := false
	for _, p := range parts[1:] {
		if len(p) == 0 {
			continue
		}
		if skip == true {
			skip = false
			continue
		}

		switch p {
		case "-o":
			skip = true
			continue
		case "-c":
			continue
		}

		if isSourceFile(p) {
			continue
		}

		p = strings.Replace(p, "\"", "\\\"", -1)
		fmt.Printf("\t\"%s\", \n", p)
	}
	fmt.Printf("]\n")

	return "", nil
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	GenereteYcmFlags(string(data))
}
