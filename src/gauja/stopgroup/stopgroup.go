package stopgroup

import (
	"sync"
)

type StopGroup interface {
	NotifyOnStop() <-chan error
	Stop(err error)
}

type stopGroup struct {
	once   *sync.Once
	result chan error
	action func()
}

func New(stopAction func()) StopGroup {
	return stopGroup{
		once:   new(sync.Once),
		result: make(chan error, 1),
		action: stopAction,
	}
}

func (sg stopGroup) NotifyOnStop() <-chan error {
	c := make(chan error)
	go func() {
		err := <-sg.result
		sg.result <- err
		c <- err
	}()
	return c
}

func (sg stopGroup) Stop(err error) {
	sg.once.Do(func() {
		sg.result <- err
		go sg.action()
	})
}
