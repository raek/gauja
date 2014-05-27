package stopgroup

import (
	"testing"
	"time"
)

var timeout = 10 * time.Second

func TestIsStoppedAfterStopCall(t *testing.T) {
	sg := New(func() {
		time.Sleep(3 * time.Second)
	})
	sg.Stop(nil)
	select {
	case <-sg.NotifyOnStop():
	case <-time.After(timeout):
		t.Fail()
	}
}

func TestIsStoppedBeforeFuncIsCalled(t *testing.T) {
	sgChan := make(chan StopGroup, 1)
	wasStopped := make(chan bool)
	sg := New(func() {
		sg := <-sgChan
		select {
		case <-sg.NotifyOnStop():
			wasStopped <- true
		default:
			wasStopped <- false
		}
	})
	sgChan <- sg
	go func() {
		sg.Stop(nil)
	}()
	if !<-wasStopped {
		t.Fail()
	}
}
