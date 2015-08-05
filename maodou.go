package maodou

import (
	"log"
	"net/http"
	"time"

	"github.com/mnhkahn/maodou/dao"
	. "github.com/mnhkahn/maodou/logs"
	"github.com/mnhkahn/maodou/models"
)

type Handler interface {
	Init()
	Start()
	Index(resp *Response)
	Detail(resp *Response)
	Result(result *models.Result)
	Config() *HandlerConfig
	Route(http_method, route string, function func(w http.ResponseWriter, req *http.Request))
	Serve(ip string, port int, graceful_timeout int)
}

type HandlerConfig struct {
	dao_name         string
	dsn              string
	cawl_every       time.Duration
	interval         time.Duration
	ip               string
	port             int
	graceful_timeout int
}

type MaoDou struct {
	Dao      dao.DaoContainer
	req      *Request
	resp     *Response
	settings *HandlerConfig
	// supervisor *supervisor.SupervisorController
	Debug bool
}

func (this *MaoDou) SetRate(times ...time.Duration) {
	if len(times) == 1 {
		this.settings.cawl_every = times[0]
	} else if len(times) == 2 {
		this.settings.cawl_every = times[0]
		this.req.Interval = times[1]
	}
}

func (this *MaoDou) SetServe(ip string, port int, graceful_timeout int) {
	this.settings.ip, this.settings.port, this.settings.graceful_timeout = ip, port, graceful_timeout
}

func (this *MaoDou) SetDao(dao_name, dsn string) {
	var err error
	this.settings.dao_name, this.settings.dsn = dao_name, dsn
	this.Dao, err = dao.NewDao(this.settings.dao_name, this.settings.dsn)
	if err != nil {
		panic(err)
	}
}

func (this *MaoDou) Init() {
	var err error
	this.settings = new(HandlerConfig)
	this.settings.cawl_every = 0
	this.settings.interval = 0
	this.settings.graceful_timeout = 1
	// this.supervisor = supervisor.NewSupervisorController()
	this.settings.dao_name = "sqlite"
	this.settings.dsn = "./maodou"
	this.req = NewRequest(0)
	this.Dao, err = dao.NewDao(this.settings.dao_name, this.settings.dsn)
	if err != nil {
		panic(err)
	}
}

func (this *MaoDou) Start() {
	log.Println("Start Method is not override.")
}

func (this *MaoDou) Cawl(paras ...interface{}) (*Response, error) {
	return this.req.Cawl(paras...)
}

func (this *MaoDou) Index(resp *Response) {
	log.Println("Start Method is not override.")
	this.Detail(nil)
}

func (this *MaoDou) Detail(resp *Response) {
	log.Println("Start Method is not override.")
	this.Result(nil)
}

func (this *MaoDou) Result(result *models.Result) {
	log.Println("Start Method is not override.")
}

func (this *MaoDou) Config() *HandlerConfig {
	return this.settings
}

func (this *MaoDou) Route(http_method, route string, function func(w http.ResponseWriter, req *http.Request)) {
	// this.supervisor.Route(http_method, route, function)
}

func (this *MaoDou) Serve(ip string, port int, graceful_timeout int) {
	// this.supervisor.Run(ip, port, graceful_timeout)
}

var APP *App

type App struct {
	handler Handler
}

func NewController(handler Handler) *App {
	app := new(App)
	app.handler = handler
	return app
}

func (this *App) Run() error {
	duration := time.Duration(0)
	if this.handler.Config() != nil && this.handler.Config().cawl_every > 0 {
		duration = this.handler.Config().cawl_every
	}
	APP.run()
	if duration > 0 {
		timer := time.NewTicker(duration)
		for {
			select {
			case <-timer.C:
				go func() {
					APP.run()
				}()
			}
		}
	}

	if this.handler.Config().port > 0 {
		this.handler.Serve(this.handler.Config().ip, this.handler.Config().port, this.handler.Config().graceful_timeout)
	}
	return nil
}

func (this *App) run() {
	ColorLog("[INFO] Start parse at %s...\n", time.Now().Format(time.RFC3339))
	this.handler.Start()
	ColorLog("[SUCC] Parse end.\n")
}

func Register(handler Handler) {
	APP = NewController(handler)
	APP.Run()
}
