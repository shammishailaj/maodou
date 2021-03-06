package maodou

import (
	"fmt"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/mnhkahn/gogogo/logger"

	"github.com/mnhkahn/maodou/request"
	"github.com/mnhkahn/maodou/request/goreq"
	"github.com/mnhkahn/maodou/request/proxy"
)

type Request struct {
	goreq.Request
	proxy    proxy.ProxyContainer
	root     string
	Interval time.Duration
}

func NewRequest(interval time.Duration) *Request {
	req := new(Request)
	req.Method = "GET"
	req.Timeout = time.Duration(30) * time.Second
	req.AddHeader("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6,zh-TW;q=0.4")
	req.AddHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Interval = interval
	// req.ShowDebug = true

	goreq.SetConnectTimeout(5 * time.Second)

	return req
}

const (
	CawlNoProxy = 0
	CawlProxy   = 1
	CawlRetry   = 2
)

func (this *Request) Cawl(paras ...interface{}) (*Response, error) {
	this.Uri = paras[0].(string)

	// Add referer
	if this.root == "" {
		this.root = this.Uri
	} else {
		this.UpdateHeader("Referer", this.root)
	}

	var p *proxy.ProxyConfig
	if this.proxy != nil && (len(paras) == 1 || (len(paras) == 2 && paras[1].(int) == CawlProxy)) {
		u := new(urlpkg.URL)
		p = this.proxy.One()
		if p.Ip != "" {
			u.Scheme = "http"
			u.Host = fmt.Sprintf("%s:%d", p.Ip, p.Port)
			this.Proxy = u.String()
		}
	}

	logger.Debug("Start to Parse:", this.Uri)

	start := time.Now()
	this.ShowDebug = true
	this.UserAgent = request.UserAgent()
	http_resp, err := this.Do()
	logger.Debugf("Cawl use %v.\n", time.Now().Sub(start))
	// 修复代理错乱的问题，需要重置代理
	this.Proxy = ""
	if err != nil {
		if len(paras) == 1 || (len(paras) == 2 && paras[1].(int) == CawlProxy) {
			this.proxy.DeleteProxy(p.Id)
		}
		logger.Warnf("Cawl Error: %s\n", err.Error())

		// Retry
		if len(paras) == 2 && paras[1].(int) == CawlRetry {
			logger.Debug("Retry...")
			this.Cawl(paras...)
		} else {
			return nil, err
		}
	}

	var resp *Response
	if http_resp.StatusCode == http.StatusOK {
		resp, err = NewResponse(http_resp.Body, this.Uri)
		if err != nil {
			if len(paras) == 2 && paras[1].(int) == CawlRetry {
				logger.Debug("Retry...")
				this.Cawl(paras...)
			} else {
				logger.Warnf("Cawl Error: %s.\n", err.Error())
				return resp, err
			}
		} else {
			logger.Debug("Cawl Success.")
		}
	} else {
		if len(paras) == 2 && paras[1].(int) == CawlRetry {
			logger.Debug("Retry...")
			this.Cawl(paras...)
		} else {
			if http_resp.StatusCode == http.StatusMovedPermanently || http_resp.StatusCode == http.StatusFound {
				logger.Debug(this.Uri, http_resp.StatusCode)
				if len(paras) == 2 {
					return this.Cawl(http_resp.Header.Get("Location"), paras[1])
				}
				return this.Cawl(http_resp.Header.Get("Location"))
			} else {
				if len(paras) == 1 || (len(paras) == 2 && paras[1].(int) == CawlProxy) {
					this.proxy.DeleteProxy(p.Id)
				}

				logger.Info("Cawl Got Status Code %d.\n", http_resp.StatusCode)
				return resp, fmt.Errorf("Cawl Got Status Code %d.", http_resp.StatusCode)
			}
		}
	}

	if this.Interval > 0 {
		time.Sleep(this.Interval)
	}
	return resp, nil
}

func (this *Request) InitProxy(proxy_name, proxy_config string) {
	this.proxy, _ = proxy.NewProxy(proxy_name, proxy_config)
	this.proxy.Init()
}

func (this *Request) DumpRequest() string {
	return this.Uri + "?" + urlpkg.Values(this.QueryString.(urlpkg.Values)).Encode()
}
