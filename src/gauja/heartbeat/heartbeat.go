package heartbeat

import (
	"errors"
	"gauja/msg"
	"gauja/stopgroup"
	"time"
)

type Durations struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var (
	ErrReadTimeout  = errors.New("heartbeat: read timeout")
	ErrWriteTimeout = errors.New("heartbeat: write timeout")
)

func Manage(upstream msg.MessageChans, sg stopgroup.StopGroup, ds Durations) (downstream msg.MessageChans) {
	downstream, myMsgs := msg.MessageChansPair()
	go func() {
		defer close(myMsgs.W)
		onStop := sg.NotifyOnStop()
		for {
			select {
			case <-onStop:
				return
			case <-time.After(ds.ReadTimeout):
				sg.Stop(ErrReadTimeout)
				return
			case msg, ok := <-upstream.R:
				if !ok {
					return
				}
				myMsgs.W <- msg
			}
		}
	}()
	go func() {
		defer close(upstream.W)
		onStop := sg.NotifyOnStop()
		for msg := range myMsgs.R {
			select {
			case <-onStop:
				return
			case <-time.After(ds.WriteTimeout):
				sg.Stop(ErrWriteTimeout)
				return
			case upstream.W <- msg:
			}
		}
	}()
	return
}
