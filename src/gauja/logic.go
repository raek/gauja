package gauja

func NewBot(fromServer <-chan Message, toServer chan<- Message) {
	go func() {
		defer close(toServer)
		send := makeSendFunc(toServer)
		send("NICK", "gauja")
		send("USER", "gauja", "0", "*", "Gauja Bot")
		for msg := range fromServer {
			switch msg.Command {
			case "PING":
				send("PONG", msg.Parameters...)
			case "001":
				send("JOIN", "#bot")
			}
		}
	}()
}

func makeSendFunc(toServer chan<- Message) func(command string, parameters ...string) {
	return func(command string, parameters ...string) {
		toServer <- MakeMessage(command, parameters...)
	}
}
