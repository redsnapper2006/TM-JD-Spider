package spider

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	//"strconv"
	"strings"
)

type JDSpider struct {
	url string
}

func NewJDSpider(u string) *JDSpider {
	return &JDSpider{url: u}
}

func (s *JDSpider) Crawl() {
	fmt.Println("begin JD spider")
	fmt.Println(s.url)
	res, err := http.Get(s.url)
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(s.url)
	if err != nil {
		log.Fatal(err)
	}
	productHtmlPath := filepath.Base(u.EscapedPath())
	productId := productHtmlPath[:strings.LastIndex(productHtmlPath, ".")]
	fmt.Println(productId)

	_ = os.Mkdir(productId, 0777)
	// if err != nil {
	//  log.Fatal(err)
	// }

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		log.Fatal(err)
	}

	// Find image
	doc.Find("#spec-list > ul > li > img").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		imgThumbUrl, isExist := s.Attr("src")
		imgUrl := strings.Replace(imgThumbUrl, "54x54", "450x450", -1)
		fmt.Printf("img %d: %s - %v\n", i, imgUrl, isExist)

		res, err := http.Get("https:" + imgUrl)
		if err != nil {
			log.Fatal(err)
		}
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		i2, err := url.Parse(imgUrl)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(filepath.Join(productId, filepath.Base(i2.EscapedPath())),
			content, 0777)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Find head
	var head string
	var descDetailUrl string
	var detailContent string
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		c, _ := s.Html()
		head = strings.Replace(strings.Replace(c, "href=\"//", "href=\"https://", -1),
			"src=\"//", "src=\"https://", -1)
		//fmt.Printf("head : %s \n", head)

		re := regexp.MustCompile("desc:\\s*'(.+?)'")
		descDetailUrl = re.FindAllStringSubmatch(head, -1)[0][1]
		fmt.Println(descDetailUrl)
		res, err := http.Get("https:" + descDetailUrl + "&callback=showdesc")
		if err != nil {
			log.Fatal(err)
		}
		content, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		re = regexp.MustCompile("\"content\":\"(.+)\"\\}\\)")
		detailContent = strings.Replace(
			strings.Replace(
				strings.Replace(
					re.FindAllStringSubmatch(string(content), -1)[0][1],
					"url(//", "url(https://", -1),
				"\\n", "", -1),
			"\\\"", "\"", -1)

		fmt.Println(detailContent)

	})

	// Find #detail
	doc.Find("#detail").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		detail, _ := s.Html()

		fmt.Printf("detail : %s \n", detail)

		re := regexp.MustCompile("<div class=\"loading-style1\"><b></b>.+</div>")
		err = ioutil.WriteFile(filepath.Join(productId, "detail.html"),
			[]byte("<html>\n  <head>"+
				head+
				"  </head>\n  <body>"+
				re.ReplaceAllString(
					strings.Replace(
						strings.Replace(detail,
							"href=\"//",
							"href=\"https://", -1),
						"src=\"//",
						"src=\"https://", -1),
					detailContent)+
				"  </body>\n</html>"), 0777)
		if err != nil {
			log.Fatal(err)
		}
	})

	// Find #guarantee
	doc.Find("#guarantee").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		guarantee, _ := s.Html()

		fmt.Printf("guarantee : %s \n", guarantee)

		err = ioutil.WriteFile(filepath.Join(productId, "guarantee.html"),
			[]byte("<html>\n  <head>"+
				head+
				"  </head>\n  <body>"+
				strings.Replace(
					strings.Replace(guarantee, "href=\"//", "href=\"https://", -1),
					"src=\"//", "src=\"https://", -1)+
				"  </body>\n</html>"), 0777)
		if err != nil {
			log.Fatal(err)
		}
	})
}
