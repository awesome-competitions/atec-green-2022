package server

import (
	"energy/config"
	"energy/log"
	"github.com/panjf2000/gnet"
)

var (
	suc  = []byte("HTTP/1.1 200 OK\nContent-Type: application/json\nContent-Length: 4\n\ntrue")
	fail = []byte("HTTP/1.1 200 OK\nContent-Type: application/json\nContent-Length: 5\n\nfalse")
)

type HttpServer struct {
	*gnet.EventServer

	addr       string
	multicore  bool
	handleFunc func(h *HttpCodec, body []byte)
}

func New(addr string, multicore bool, handler func(hc *HttpCodec, body []byte)) *HttpServer {
	return &HttpServer{
		addr:       addr,
		multicore:  multicore,
		handleFunc: handler,
	}
}

type HttpCodec struct {
	parser *HTTPParser
	buf    []byte
	status int
}

func (hc *HttpCodec) Suc() {
	hc.buf = append(hc.buf, suc...)
}

func (hc *HttpCodec) Fail() {
	hc.buf = append(hc.buf, fail...)
}

func (hc *HttpCodec) Path() []byte {
	return hc.parser.Path
}

func (hc *HttpCodec) Method() []byte {
	return hc.parser.Method
}

func (hs *HttpServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Infof("HTTP server is listening on %s (multi-cores: %t, event-loops: %d)",
		srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (hs *HttpServer) OnOpened(c gnet.Conn) ([]byte, gnet.Action) {
	c.SetContext(&HttpCodec{
		parser: NewHTTPParser(),
	})
	return nil, gnet.None
}

func (hs *HttpServer) React(data []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	hc := c.Context().(*HttpCodec)
	hc.parser.Reset()
	err := hc.parser.Parse(data)
	if err != nil {
		return []byte("500 Error"), gnet.Close
	}
	hs.handleFunc(hc, nil)
	out = hc.buf
	hc.buf = hc.buf[:0]
	return
}

func (hs *HttpServer) Run() error {
	return gnet.Serve(hs, hs.addr, gnet.WithMulticore(hs.multicore), gnet.WithNumEventLoop(config.GNetEventLoopNum))
}
