package stopgroup

import "testing"
import "time"

func TestIsStoppedAfterStopCall(t *testing.T) {
	sg := New(func() {
		time.Sleep(3 * time.Second)
	})
	sg.Stop()
	if !IsStopped(sg) {
		t.Fail()
	}
}

func TestIsStoppedBeforeFuncIsCalled(t *testing.T) {
	sgChan := make(chan StopGroup, 1)
	wasStopped := make(chan bool)
	sg := New(func() {
		sg := <-sgChan
		wasStopped <- IsStopped(sg)
	})
	sgChan <- sg
	go func() {
		sg.Stop()
	}()
	if !<-wasStopped {
		t.Fail()
	}
}
