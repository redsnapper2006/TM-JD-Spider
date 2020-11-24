package spider

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	//"golang.org/x/net/html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	// "bytes"
	//"strconv"
	"encoding/json"
	"strings"
)

type TMSpider struct {
	url string
}

func NewTMSpider(u string) *TMSpider {
	return &TMSpider{url: u}
}

func (s *TMSpider) Crawl() {
	fmt.Println("begin TM spider")
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
	// fmt.Println(string(content))

	u, err := url.Parse(s.url)
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	productId := q["id"][0]
	fmt.Println(q["id"])

	_ = os.Mkdir(productId, 0777)
	// if err != nil {
	//  log.Fatal(err)
	// }

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		log.Fatal(err)
	}

	// Find head
	var head string
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		c, _ := s.Html()
		head = strings.Replace(strings.Replace(c, "href=\"//", "href=\"https://", -1),
			"src=\"//", "src=\"https://", -1)
		//fmt.Printf("head : %s \n", head)
	})

	// Find httpsDescUrl & fetchDcUrl
	var httpsDescUrl, fetchDcUrl string
	re := regexp.MustCompile("\"httpsDescUrl\":\"(.+?)\"")
	httpsDescUrl = re.FindAllStringSubmatch(string(content), -1)[0][1]

	re = regexp.MustCompile("\"fetchDcUrl\":\"(.+?)\"")
	fetchDcUrl = re.FindAllStringSubmatch(string(content), -1)[0][1]
	fmt.Println(httpsDescUrl, " + ", fetchDcUrl)

	// Find image
	doc.Find("#J_UlThumb > li > a > img").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		imgThumbUrl, isExist := s.Attr("src")

		imgUrl := strings.Replace(imgThumbUrl, "60x60", "430x430", -1)
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

	// Find #attributes
	doc.Find("#attributes").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		detail, _ := s.Html()

		fmt.Printf("attributes : %s \n", detail)

		err = ioutil.WriteFile(filepath.Join(productId, "attributes.html"),
			[]byte("<html>\n  <head>"+
				head+
				"  </head>\n  <body>"+
				strings.Replace(
					strings.Replace(detail,
						"href=\"//",
						"href=\"https://", -1),
					"src=\"//",
					"src=\"https://", -1)+
				"  </body>\n</html>"), 0777)
		if err != nil {
			log.Fatal(err)
		}
	})

	var dcTopRight string
	// Find #J_DcTopRightWrap
	doc.Find("#J_DcTopRightWrap").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		detail, _ := s.Html()

		fmt.Printf("dcTopRight : %s \n", detail)
		dcTopRight = string(detail)
	})

	var description string
	// Find #description
	doc.Find("#description").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		detail, _ := s.Html()

		fmt.Printf("description : %s \n", detail)
		description = string(detail)
	})

	var dcBottomRight string
	// Find #J_DcBottomRightWrap
	doc.Find("#J_DcBottomRightWrap").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		detail, _ := s.Html()

		fmt.Printf("dcBottomRight : %s \n", detail)
		dcBottomRight = string(detail)
	})

	// Fetch asynchtm
	fmt.Println("https:" + fetchDcUrl)
	res, err = http.Get("https:" + fetchDcUrl)
	if err != nil {
		log.Fatal(err)
	}
	dcContent, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("https:" + httpsDescUrl)
	res, err = http.Get("https:" + httpsDescUrl)
	if err != nil {
		log.Fatal(err)
	}
	descContent, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(descContent))

	reLF := regexp.MustCompile(`\\r`)
	reCR := regexp.MustCompile(`\\n`)
	reTab := regexp.MustCompile(`\t`)
	shopDc :=
		reTab.ReplaceAllString(
			reCR.ReplaceAllString(
				reLF.ReplaceAllString(
					strings.Replace(
						strings.Replace(string(dcContent),
							"var SHOP_DC = ", "", -1),
						"};", "}", -1), ""), ""), "")

	//fmt.Printf("asynchtm : %s\n", shopDc)

	//shopDc = `{"topRight":"a", "bottomRight": "b"}`

	var sa map[string]interface{}
	err = json.Unmarshal([]byte(shopDc), &sa)
	if err != nil {
		fmt.Println("error:", err)
	}
	// fmt.Printf("%+v", sa)

	// dcTopRightNode, err := html.Parse(strings.NewReader(dcTopRight))
	// if err != nil {
	//     log.Fatal(err)
	// }
	// dcTopRightChildNode, err := html.Parse(strings.NewReader(sa["topRight"].(string)))
	// if err != nil {
	//     log.Fatal(err)
	// }
	// dcTopRightNode.InsertBefore(dcTopRightChildNode, nil)

	// var b bytes.Buffer
	// err = html.Render(&b, dcTopRightNode)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// fmt.Println(b.String())

	reTopRight := regexp.MustCompile(`id="J_DcTopRight"(.+?)><`)
	//fmt.Println(reTopRight.ReplaceAllString(dcTopRight, `id="J_DcTopRight"$1>`+sa["topRight"].(string)+"<"))

	reDescription := regexp.MustCompile(`<div class="content ke-post">\n.+</div>`)
	// fmt.Println(reDescription.ReplaceAllString(description, `<div class="content ke-post">`+
	// 	strings.Replace(
	// 		strings.Replace(string(descContent), "var desc='", "", -1), "';", "'", -1)+"</div>"))

	reBottomRight := regexp.MustCompile(`id="J_DcBottomRight"(.+?)><`)
	fmt.Println(reBottomRight.ReplaceAllString(dcBottomRight, `id="J_DcBottomRight"$1>`+sa["bottomRight"].(string)+"<"))

	err = ioutil.WriteFile(filepath.Join(productId, "description.html"),
		[]byte("<html>\n  <head>"+
			head+
			"  </head>\n  <body>"+
			strings.Replace(
				strings.Replace(reTopRight.ReplaceAllString(dcTopRight,
					`id="J_DcTopRight"$1>`+
						sa["topRight"].(string)+"<"),
					"href=\"//",
					"href=\"https://", -1),
				"src=\"//",
				"src=\"https://", -1)+
			"\n"+
			strings.Replace(
				strings.Replace(reDescription.ReplaceAllString(description,
					`<div class="content ke-post">`+
						strings.Replace(
							strings.Replace(string(descContent),
								"var desc='", "", -1),
							"';", "'", -1)+
						"</div>"),
					"href=\"//",
					"href=\"https://", -1),
				"src=\"//",
				"src=\"https://", -1)+
			"\n"+
			strings.Replace(
				strings.Replace(
					strings.Replace(reBottomRight.ReplaceAllString(dcBottomRight,
						`id="J_DcBottomRight"$1>`+
							sa["bottomRight"].(string)+"<"),
						"href=\"//",
						"href=\"https://", -1),
					"src=\"//",
					"src=\"https://", -1),
				`src="https://assets.alicdn.com/s.gif" data-ks-lazyload="`, `src="https:`, -1)+
			"  </body>\n</html>"), 0777)

	if err != nil {
		log.Fatal(err)
	}
}
