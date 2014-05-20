package lineio

import (
	"bufio"
	"gauja/stopgroup"
	"io"
)

type LineChans struct {
	R <-chan string
	W chan<- string
}

func LineChansPair() (a, b LineChans) {
	aToB := make(chan string)
	bToA := make(chan string)
	a = LineChans{R: bToA, W: aToB}
	b = LineChans{R: aToB, W: bToA}
	return
}

type lineIo struct {
	io.ReadWriteCloser
	stopgroup.StopGroup
	LineChans
}

func Manage(rwc io.ReadWriteCloser) (lines LineChans, sg stopgroup.StopGroup) {
	sg = stopgroup.New(func() {
		rwc.Close()
	})
	lines, myLines := LineChansPair()
	lio := lineIo{rwc, sg, myLines}
	go lio.manageReads()
	go lio.manageWrites()
	return
}

func (lio lineIo) manageReads() {
	defer lio.Stop()
	defer close(lio.W)
	s := bufio.NewScanner(lio)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		lio.W <- s.Text()
	}
}

func (lio lineIo) manageWrites() {
	defer lio.Stop()
	w := bufio.NewWriter(lio)
	for {
		select {
		case <-lio.Stopped():
			return
		case line, ok := <-lio.R:
			if !ok {
				return
			}
			w.WriteString(line)
			w.WriteString("\r\n")
			err := w.Flush()
			if err != nil {
				return
			}
		}
	}
}
