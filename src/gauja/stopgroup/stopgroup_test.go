package stopgroup

import (
	"testing"
	"time"
)

func TestIsStoppedAfterStopCall(t *testing.T) {
	sg := New(func() {
		time.Sleep(3 * time.Second)
	})
	sg.Stop(nil)
	_, stopped := sg.SampleStopState()
	if !stopped {
		t.Fail()
	}
}

func TestIsStoppedBeforeFuncIsCalled(t *testing.T) {
	sgChan := make(chan StopGroup, 1)
	wasStopped := make(chan bool)
	sg := New(func() {
		sg := <-sgChan
		_, stopped := sg.SampleStopState()
		wasStopped <- stopped
	})
	sgChan <- sg
	go func() {
		sg.Stop(nil)
	}()
	if !<-wasStopped {
		t.Fail()
	}
}
