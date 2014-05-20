package main

import (
	"fmt"
	"gauja/heartbeat"
	"gauja/lineio"
	"gauja/linelog"
	"gauja/msg"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var heartbeatPeriod = 10 * time.Second
var connectTimeout = heartbeatPeriod
var pongPeriod = heartbeatPeriod / 2

func main() {
	interrupted := make(chan os.Signal)
	defer signal.Stop(interrupted)
	signal.Notify(interrupted, os.Interrupt)
	ds := heartbeat.Durations{
		ReadTimeout:  heartbeatPeriod,
		WriteTimeout: heartbeatPeriod / 10,
	}
	address := "irc.raek.se:6667"
	netConn, err := net.DialTimeout("tcp", address, connectTimeout)
	if err != nil {
		log.Fatal(err)
	}
	lines, sg := lineio.Manage(netConn)
	lines = linelog.Manage(lines)
	msgs := msg.Manage(lines)
	msgs = heartbeat.Manage(msgs, sg, ds)
	go func() {
		send := func(command string, parameters ...string) {
			msgs.W <- msg.MakeMessage(command, parameters...)
		}
		defer close(msgs.W)
		send("NICK", "gauja")
		send("USER", "gauja", "0", "*", "Gauja Bot")
		for _ = range msgs.R {
		}
	}()
	select {
	case <-interrupted:
		fmt.Println("Interrupted")
		sg.Stop()
	case <-sg.Stopped():
	}
	fmt.Println("Shutting down...")
	time.Sleep(3 * time.Second)
}
