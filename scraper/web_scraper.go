package scraper

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

type Scraper interface {
	GetAll() ([]Move, error)
	GetByCode(code string) (Move, error)
}

type Move struct {
	Code   string
	Author string
	Moves  string
}

type WebScraper struct {
	ecoCodeDataURL string
	cache          map[string]Move
	list           []Move
	ttl            *time.Timer
}

func NewWebScraper(url string) *WebScraper {
	return &WebScraper{
		ecoCodeDataURL: url,
		cache:          make(map[string]Move),
	}
}

func (ws *WebScraper) GetAll() ([]Move, error) {
	if len(ws.list) <= 0 {
		ws.crawl()
	}
	return ws.list, nil
}

func (ws *WebScraper) GetByCode(code string) (Move, error) {
	if len(ws.cache) <= 0 {
		ws.crawl()
	}
	return ws.cache[code], nil
}

func (ws *WebScraper) crawl() error {
	htmlStr, err := ws.getHTML([]byte(ws.ecoCodeDataURL))
	if err != nil {
		return err
	}

	return ws.getCodesFromHTMLString(htmlStr)
}

// GetHTML returns html string
func (ws *WebScraper) getHTML(url []byte) ([]byte, error) {
	res, err := http.Get(string(url))
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return b, err
}

// Parse html body and get the chess moves
func (ws *WebScraper) getCodesFromHTMLString(body []byte) error {
	doc, err := html.Parse(bytes.NewBuffer(body))
	if err != nil {
		log.Fatal(err)
		return err
	}

	count := 0
	var f func(*html.Node)
	f = func(n *html.Node) {
		if count >= 100 {
			return
		}
		if n.Type == html.ElementNode && n.Data == "tr" {
			k := n.FirstChild.FirstChild.FirstChild.Data
			cv := n.FirstChild.NextSibling.FirstChild
			author := cv.FirstChild.FirstChild.Data
			code := cv.LastChild.FirstChild.Data
			m := Move{k, author, code}
			ws.cache[k] = m
			ws.list = append(ws.list, m)
			count++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// Set cache TTL and go-routine to clear cache on expiry
	ws.ttl = time.NewTimer(180 * time.Second)
	go ws.reset()

	return nil
}

func (ws *WebScraper) reset() {
	<-ws.ttl.C
	log.Println("TTL expired. Clearing cache...")
	ws.cache = make(map[string]Move)
	ws.list = make([]Move, 0)
}
