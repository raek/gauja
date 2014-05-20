package linelog

import (
	"fmt"
	"gauja/lineio"
)

func Manage(upstream lineio.LineChans) (downstream lineio.LineChans) {
	downstream, myLines := lineio.LineChansPair()
	go func() {
		defer close(myLines.W)
		for line := range upstream.R {
			fmt.Println("--> " + line)
			myLines.W <- line
		}
	}()
	go func() {
		defer close(upstream.W)
		for line := range myLines.R {
			fmt.Println("<-- " + line)
			upstream.W <- line
		}
	}()
	return
}
