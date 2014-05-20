package main

import (
	"gauja"
	"log"
	"os"
	"os/signal"
	"time"
)

var timeoutSettings = gauja.TimeoutSettings{
	ConnectTimeout: 3 * time.Minute,
	ReadTimeout:    10 * time.Minute,
	WriteTimeout:   1 * time.Minute,
}

func main() {
	handleLink()
}

func handleLink() {
	interrupted := make(chan os.Signal)
	signal.Notify(interrupted, os.Interrupt)
	for {
		handleConn(interrupted)
		return
	}
}

func handleConn(interrupted <-chan os.Signal) {
	conn, err := gauja.Connect("irc.raek.se:6667", timeoutSettings)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case <-interrupted:
			conn.Stop()
		case <-conn.Stopped:
			println("bye")
			return
		}
	}
}
