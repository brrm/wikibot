package main

func (c comment) handler(bot redditBot) {
	if c.author != bot.Username {
		query := getQuery(c.text)
		if query == "" {
			return
		}
		if title, summary, err := getWiki(query); err == nil {
			bot.postComment(c.id, "#"+title+"\n"+summary)
		}
	}
}

func main() {
	bot := loadBot()
	if err := bot.auth(); err != nil {
		panic(err)
	}

	c := make(chan redditItem)
	go bot.streamComments(c)
	bot.handlerManager(c)
}
