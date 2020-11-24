package main

import (
	"rs.pm/spider"
)

const (
	JD_URL = "https://item.jd.com/7434988.html?q=a"
	TM_URL = "https://detail.tmall.com/item.htm?id=556923025304&skuId=3999719324886"
)

func main() {

	var s spider.Spider
	s = spider.NewTMSpider(TM_URL)
	s.Crawl()
}
