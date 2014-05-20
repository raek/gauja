package lineio

import (
	"io"
	"io/ioutil"
	"testing"
	"time"
)

var timeout = 10 * time.Second

type PipeRwc struct {
	io.Reader
	io.Writer
	ReadCloser  io.Closer
	WriteCloser io.Closer
}

func (p PipeRwc) Close() error {
	p.ReadCloser.Close()
	p.WriteCloser.Close()
	return nil
}

func MakePipeRwc() (rwcA, rwcB PipeRwc) {
	readerA, writerB := io.Pipe()
	readerB, writerA := io.Pipe()
	rwcA = PipeRwc{readerA, writerA, readerA, writerA}
	rwcB = PipeRwc{readerB, writerB, readerB, writerB}
	return
}

func TestIncomingChanIsClosedWhenStopped(t *testing.T) {
	rwcA, _ := MakePipeRwc()
	lines, sg := Manage(rwcA)
	sg.Stop()
	select {
	case _, ok := <-lines.R:
		if ok {
			t.Fail()
		}
	case <-time.After(timeout):
		t.Error("chan was never closed")
	}
}

func TestReadLines(t *testing.T) {
	rwcA, rwcB := MakePipeRwc()
	lines, sg := Manage(rwcA)
	defer sg.Stop()
	go func() {
		io.WriteString(rwcB, "a\nb\nc\n")
		rwcB.Close()
	}()
	line1 := <-lines.R
	line2 := <-lines.R
	line3 := <-lines.R
	_, ok := <-lines.R
	if line1 != "a" ||
		line2 != "b" ||
		line3 != "c" ||
		ok != false {
		t.Fail()
	}
}

func TestWriteLines(t *testing.T) {
	rwcA, rwcB := MakePipeRwc()
	lines, sg := Manage(rwcA)
	defer sg.Stop()
	go func() {
		lines.W <- "a"
		lines.W <- "b"
		lines.W <- "c"
		close(lines.W)
	}()
	bytes, err := ioutil.ReadAll(rwcB)
	if err != nil {
		t.Error(err)
		return
	}
	if string(bytes) != "a\r\nb\r\nc\r\n" {
		t.Fail()
	}
}
