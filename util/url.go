package util

import (
	"log"
	"net/url"
	"strings"
)

func JoinURL(p1, p2 string) string {
	p1HasSlash := strings.HasSuffix(p1, "/")
	p2HasSlash := strings.HasPrefix(p2, "/")

	if p1HasSlash && p2HasSlash {
		return p1 + p2[1:]
	} else if !p1HasSlash && !p2HasSlash {
		return p1 + "/" + p2
	} else {
		return p1 + p2
	}
}

func FormURL(server, path string) string {
	if len(server) == 0 {
		log.Fatal("invalid server address")
	}

	if !strings.HasPrefix(server, "http://") {
		server = "http://" + server
	}

	return JoinURL(server, path)
}

func FormRedirectUrl(old string, redirect string) (string, error) {
	oldParsed, err := url.Parse(old)
	if err != nil {
		return "", err
	}

	redirectParsed, err := url.Parse(redirect)
	if err != nil {
		return "", err
	}

	oldParsed.Host = redirectParsed.Host
	return oldParsed.String(), nil
}
