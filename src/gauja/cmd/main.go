package main

import (
	"bufio"
	"fmt"
	"gauja"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	readTimeout  = 10 * time.Minute
	writeTimeout = 1 * time.Minute
)

func main() {
	err := do()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func do() error {
	conn, err := connect("irc.raek.se:6667")
	if err != nil {
		return err
	}
	handle("FesternNet", conn)
	return nil
}

func connect(address string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type conn struct {
	NetworkName   string
	NetConn       net.Conn
	FromServer    chan<- gauja.Message
	ToServer      <-chan gauja.Message
	ControlLogger *log.Logger
	ReadLogger    *log.Logger
	WriteLogger   *log.Logger
	Terminated    <-chan Void
	Terminate     func()
}

func MakeConn(networkName string, netConn net.Conn, fromServer chan<- gauja.Message, toServer <-chan gauja.Message) conn {
	controlLogger := log.New(os.Stdout, fmt.Sprintf("%s -!- ", networkName), 0)
	readLogger := log.New(os.Stdout, fmt.Sprintf("%s --> ", networkName), 0)
	writeLogger := log.New(os.Stdout, fmt.Sprintf("%s <-- ", networkName), 0)
	terminated := make(chan Void)
	requestTermination := make(chan Unit)
	terminate := func() {
		select {
		case requestTermination <- Unit{}:
		default:
		}
	}
	go func() {
		<-requestTermination
		defer close(terminated)
		if err := netConn.Close(); err != nil {
			controlLogger.Printf("Error when closing connection: %v", err)
		}
	}()
	return conn{networkName, netConn, fromServer, toServer, controlLogger, readLogger, writeLogger, terminated, terminate}
}

func (conn conn) isTerminated() bool {
	select {
	case <-conn.Terminated:
		return true
	default:
		return false
	}
}

type Unit struct{}

type Void struct{}

func handle(networkName string, netConn net.Conn) {
	var wg sync.WaitGroup
	fromServer := make(chan gauja.Message)
	toServer := make(chan gauja.Message)
	conn := MakeConn(networkName, netConn, fromServer, toServer)
	conn.ControlLogger.Print("Connected to server")
	gauja.NewBot(fromServer, toServer)
	wg.Add(3)
	go func() {
		defer wg.Done()
		conn.handleFromServer()
	}()
	go func() {
		defer wg.Done()
		conn.handleToServer()
	}()
	go func() {
		defer wg.Done()
		<-conn.Terminated
	}()
	wg.Wait()
	conn.ControlLogger.Print("Disconnected from server")
}

func (conn conn) handleFromServer() {
	logger := conn.ReadLogger
	defer close(conn.FromServer)
	defer logger.Print("Done")
	scanner := bufio.NewScanner(conn.NetConn)
	scanner.Split(bufio.ScanLines)
	conn.NetConn.SetReadDeadline(time.Now().Add(readTimeout))
	for scanner.Scan() {
		conn.NetConn.SetReadDeadline(time.Now().Add(readTimeout))
		line := scanner.Text()
		msg := gauja.ParseMessage(line)
		logger.Print(msg)
		conn.FromServer <- msg
	}
	if err := scanner.Err(); err != nil {
		if conn.isTerminated() {
			logger.Print("Connection was terminated while reading")
			return
		}
		if isTimeout(err) {
			logger.Printf("Timeout (%v)", readTimeout)
		} else {
			logger.Printf("Error when reading: %v", err)
		}
	} else {
		logger.Print("Connection closed by server")
	}
	terminateConnection(conn, logger)
}

func (conn conn) handleToServer() {
	logger := conn.WriteLogger
	defer logger.Print("Done")
	w := bufio.NewWriter(conn.NetConn)
	for {
		select {
		case <-conn.Terminated:
			logger.Print("Connection was terminated while selecting")
			return
		case msg, ok := <-conn.ToServer:
			if !ok {
				logger.Print("ToServer closed!")
				terminateConnection(conn, logger)
				return
			}
			logger.Print(msg)
			conn.NetConn.SetWriteDeadline(time.Now().Add(writeTimeout))
			_, err := w.WriteString(msg.String())
			_, err = w.WriteString("\r\n")
			err = w.Flush()
			if err != nil {
				if conn.isTerminated() {
					logger.Print("Connection was terminated while writing")
					return
				}
				if isTimeout(err) {
					logger.Printf("Timeout (%v)", writeTimeout)
				} else {
					logger.Printf("Error when writing: %v", err)
				}
				terminateConnection(conn, logger)
				return
			}
		}
	}
}

func terminateConnection(conn conn, logger *log.Logger) {
	logger.Print("Terminating connection")
	conn.Terminate()
}

func isTimeout(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	} else {
		return false
	}
}
