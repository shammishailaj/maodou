package proxy

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"math/rand"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/franela/goreq"
)

type XiciConfig struct {
	MaxCawlCnt int    `json:"max_cawl_cnt"`
	Cnt        int    `json:"cnt"`
	MinCnt     int    `json:"min_cnt"`
	Root       string `json:"root"`
}

type XiciProxy struct {
}

func (this *XiciProxy) NewProxyImpl(dsn string) (ProxyContainer, error) {
	d := new(XiciProxyContainer)
	config := new(XiciConfig)
	err := json.Unmarshal([]byte(dsn), config)
	d.config = config
	if err != nil {
		return d, fmt.Errorf("Config for xici is error: %v", err)
	}

	return d, nil
}

type XiciProxyContainer struct {
	config  *XiciConfig
	req     goreq.Request
	proxies []*ProxyConfig
}

func (this *XiciProxyContainer) Init() {
	this.init()
}

func (this *XiciProxyContainer) init() {
	resp, err := goreq.Request{Uri: "http://www.xici.net.co/"}.Do()
	if err == nil {
		raw_document, _ := html.Parse(resp.Body)
		document := goquery.NewDocumentFromNode(raw_document)
		document.Find(`#ip_list > tbody > tr`).Each(func(i int, s *goquery.Selection) {
			if i > 1 && len(this.proxies) < this.config.Cnt && s.Children().Length() == 7 {
				p := new(ProxyConfig)
				p.Ip = s.Children().Get(1).FirstChild.Data
				p.Port, _ = strconv.Atoi(s.Children().Get(2).FirstChild.Data)
				if s.Children().Get(3).FirstChild != nil {
					p.Location = s.Children().Get(3).FirstChild.Data
				}
				p.Safe = s.Children().Get(4).FirstChild.Data == "高匿"
				p.Schema = s.Children().Get(5).FirstChild.Data
				p.VerifyTime = s.Children().Get(6).FirstChild.Data
				if this.TestProxy(p) {
					p.Id = len(this.proxies)
					this.proxies = append(this.proxies, p)
				}
			}
		})
	} else {
		log.Println(err)
	}
}

func (this *XiciProxyContainer) One() *ProxyConfig {
	if len(this.proxies) == 0 {
		return new(ProxyConfig)
	}
	i := rand.Intn(len(this.proxies))
	p := this.proxies[i]
	p.Cnt++

	if p.Cnt >= this.config.MaxCawlCnt {
		this.DeleteProxy(i)
	}
	return p
}

func (this *XiciProxyContainer) Len() int {
	return len(this.proxies)
}

func (this *XiciProxyContainer) DeleteProxy(i int) {
	this.proxies = append(this.proxies[:i], this.proxies[i+1:]...)
	if len(this.proxies) <= this.config.MinCnt {
		this.Init()
	}
}

// true means proxy is OK
func (this *XiciProxyContainer) TestProxy(p *ProxyConfig) bool {
	_, err := goreq.Request{Uri: this.config.Root}.Do()
	if err != nil {
		return false
	}
	return true
}

func init() {
	Register("xici", &XiciProxy{})
}