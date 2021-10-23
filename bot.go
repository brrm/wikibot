package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/buger/jsonparser"
)

type redditBot struct {
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
	UserAgent    string

	Subreddits []string

	token string
}

func loadBot() redditBot {
	dat, err := os.ReadFile("config.toml")
	if err != nil {
		panic(err)
	}
	bot := redditBot{}
	if _, err := toml.Decode(string(dat), &bot); err != nil {
		panic(err)
	}
	return bot
}

func (bot *redditBot) auth() error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token?grant_type=password&username="+bot.Username+"&password="+bot.Password, nil)
	req.SetBasicAuth(bot.ClientID, bot.ClientSecret)
	req.Header.Set("User-Agent", bot.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	token, _, _, err := jsonparser.Get(body, "access_token")
	if err != nil {
		return err
	}
	bot.token = string(token)
	return nil
}

func (bot redditBot) postComment(parentId string, text string) error {
	params := url.Values{}
	params.Add("text", text)
	params.Add("thing_id", parentId)
	uri := "https://oauth.reddit.com/api/comment?" + params.Encode()

	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+bot.token)
	req.Header.Set("User-Agent", bot.UserAgent)

	if _, err = client.Do(req); err != nil {
		return err
	}

	return nil
}

type redditItem interface {
	handler(redditBot)
}

type comment struct {
	author string
	id     string
	text   string
}

func (bot redditBot) streamComments(c chan redditItem) {
	uri := "https://oauth.reddit.com/r/"
	for i, subreddit := range bot.Subreddits {
		uri += subreddit
		if i != len(bot.Subreddits)-1 {
			uri += "+"
		}
	}
	uri += "/comments?limit=100?sort=new"

	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+bot.token)
	req.Header.Set("User-Agent", bot.UserAgent)

	lastId := ""
	for {
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		duplicates := false
		tmpId := ""
		jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if !duplicates {
				author, _, _, _ := jsonparser.Get(value, "data", "author")
				id, _, _, _ := jsonparser.Get(value, "data", "name")
				text, _, _, _ := jsonparser.Get(value, "data", "body")
				if tmpId == "" {
					tmpId = string(id)
				}
				if string(id) == lastId {
					duplicates = true
				} else {
					c <- comment{author: string(author), id: string(id), text: string(text)}
				}
			}
		}, "data", "children")
		lastId = tmpId
		time.Sleep(500 * time.Millisecond) // Avoid going 2 requests per second rate limit
	}
}

func (bot redditBot) handlerManager(c chan redditItem) {
	for {
		item := <-c
		go item.handler(bot)
	}
}
