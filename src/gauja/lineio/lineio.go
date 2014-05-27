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
	go doAndStop(lio, lio.manageReads)
	go doAndStop(lio, lio.manageWrites)
	return
}

func doAndStop(sg stopgroup.StopGroup, f func()) {
	defer func() {
		v := recover()
		err, ok := v.(error)
		if ok {
			sg.Stop(err)
		} else {
			sg.Stop(nil)
		}
	}()
	f()
}

func (lio lineIo) manageReads() {
	defer lio.Stop(nil)
	defer close(lio.W)
	s := bufio.NewScanner(lio)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		lio.W <- s.Text()
	}
	err := s.Err()
	if err != nil {
		lio.Stop(err)
	} else {
		lio.Stop(io.EOF)
	}
}

func (lio lineIo) manageWrites() {
	defer lio.Stop(nil)
	w := bufio.NewWriter(lio)
	onStop := lio.NotifyOnStop()
	for {
		select {
		case <-onStop:
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
