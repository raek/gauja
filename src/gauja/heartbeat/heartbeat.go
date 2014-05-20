package heartbeat

import (
	"gauja/msg"
	"gauja/stopgroup"
	"time"
)

type Durations struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Manage(upstream msg.MessageChans, sg stopgroup.StopGroup, ds Durations) (downstream msg.MessageChans) {
	downstream, myMsgs := msg.MessageChansPair()
	go func() {
		defer close(myMsgs.W)
		for {
			select {
			case <-sg.Stopped():
				return
			case <-time.After(ds.ReadTimeout):
				sg.Stop()
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
		for msg := range myMsgs.R {
			select {
			case <-sg.Stopped():
				return
			case <-time.After(ds.WriteTimeout):
				sg.Stop()
				return
			case upstream.W <- msg:
			}
		}
	}()
	return
}
