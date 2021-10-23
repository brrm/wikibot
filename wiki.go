package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
)

func getQuery(text string) string {
	urls := FindURL().FindStringSubmatch(text)
	for _, url := range urls {
		if strings.Contains(url, "#") { //ignore section urls
			continue
		}
		if strings.Contains(url, "en.wikipedia.org/wiki/") {
			return strings.TrimSpace(strings.Split(url, "en.wikipedia.org/wiki/")[1])
		}
		if strings.Contains(url, "www.wikipedia.org/wiki/") {
			return strings.TrimSpace(strings.Split(url, "www.wikipedia.org/wiki/")[1])
		}
	}
	return ""
}

func getWiki(query string) (string, string, error) {
	resp, err := http.Get("https://en.wikipedia.org/api/rest_v1/page/summary/" + query)
	if err != nil {
		return "", "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	title, _, _, err := jsonparser.Get(body, "title")
	if err != nil {
		return "", "", err
	}
	summary, _, _, err := jsonparser.Get(body, "extract")
	if err != nil {
		return "", "", err
	}
	return string(title), string(summary), nil
}
