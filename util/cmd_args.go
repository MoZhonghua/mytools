package util

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"
)

func SplitListArg(arg string, seperator string) []string {
	results := make([]string, 0)
	for _, f := range strings.Split(arg, seperator) {
		normalized := strings.TrimSpace(f)
		if normalized != "" {
			results = append(results, f)
		}
	}

	return results
}

func HasFlag(name string) bool {
	has := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			has = true
		}
	})
	return has
}

func LoadJsonConfig(path string, c interface{}) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func DumpConfig(c interface{}) string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func FileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}
