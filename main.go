package main

import (
	"energy/db"
	"energy/handler"
	"energy/log"
	"energy/server"
	_ "go.uber.org/automaxprocs"
)

func main() {
	//d, err := db.New("192.168.1.2", 3306, "root", "root")
	d, err := db.New("127.0.0.1", 3306, "root", "111111")
	if err != nil {
		log.Infof("db.New err: %v", err)
		return
	}

	h := handler.New(d)
	err = h.Init()
	if err != nil {
		log.Infof("h.Init err: %v", err)
		return
	}

	s := server.New("tcp://0.0.0.0:8080", true, h.Handle)
	log.Infof("server stopped: %v", s.Run())
}
