package msg

import (
	"bytes"
	"gauja/lineio"
	"strings"
)

type Message struct {
	Sender     Sender
	Command    string
	Parameters []string
}

type Sender struct {
	Nick  string
	Login string
	Host  string
}

func MakeMessage(command string, parameters ...string) Message {
	return Message{
		Sender:     Sender{},
		Command:    command,
		Parameters: parameters,
	}
}

// Formatting: Message -> string

func (msg Message) String() string {
	var buffer bytes.Buffer
	if msg.Sender.Nick != "" {
		buffer.WriteString(":")
		buffer.WriteString(msg.Sender.Nick)
		if msg.Sender.Login != "" {
			buffer.WriteString("!")
			buffer.WriteString(msg.Sender.Login)
		}
		if msg.Sender.Host != "" {
			buffer.WriteString("@")
			buffer.WriteString(msg.Sender.Host)
		}
		buffer.WriteString(" ")
	}
	buffer.WriteString(msg.Command)
	n := len(msg.Parameters)
	if n > 0 {
		for _, parameter := range msg.Parameters[0 : n-1] {
			buffer.WriteString(" ")
			buffer.WriteString(parameter)
		}
		buffer.WriteString(" :")
		buffer.WriteString(msg.Parameters[n-1])
	}
	return buffer.String()
}

// Parsing: string -> Message

type parseState struct {
	input string
	pos   int
}

func ParseMessage(line string) Message {
	p := &parseState{
		input: line,
		pos:   0,
	}
	return Message{
		Sender:     p.parseSender(),
		Command:    p.parseCommand(),
		Parameters: p.parseParameters(),
	}
}

func (p *parseState) parseSender() Sender {
	if p.tryParseColon() {
		sender := p.parseWord()
		nickAndLogin, host := splitOffOptionalSuffix(sender, "@")
		nick, login := splitOffOptionalSuffix(nickAndLogin, "!")
		return Sender{nick, login, host}
	}
	return Sender{}
}

func splitOffOptionalSuffix(s, del string) (string, string) {
	i := strings.Index(s, del)
	if i == -1 {
		return s, ""
	} else {
		return s[:i], s[i+1:]
	}
}

func (p *parseState) parseCommand() string {
	return p.parseWord()
}

func (p *parseState) parseParameters() []string {
	var result []string
	var i int = 0
	for p.hasInput() {
		if i >= 1000 {
			panic("Infinite loop in parseParameters")
		}
		i++
		if p.tryParseColon() {
			result = append(result, p.parseRest())
			break
		} else {
			result = append(result, p.parseWord())
		}
	}
	return result
}

func (p *parseState) tryParseColon() bool {
	if p.hasInput() && p.input[p.pos] == ':' {
		p.pos++
		return true
	} else {
		return false
	}
}

func (p *parseState) parseWord() string {
	end := strings.Index(p.input[p.pos:], " ")
	if end == -1 {
		result := p.input[p.pos:]
		p.pos = len(p.input)
		return result
	}
	result := p.input[p.pos : p.pos+end]
	p.pos = p.pos + end + 1
	return result
}

func (p *parseState) hasInput() bool {
	return p.pos < len(p.input)
}

func (p *parseState) parseRest() string {
	result := p.input[p.pos:]
	p.pos = len(p.input)
	return result
}

// Channels

type MessageChans struct {
	R <-chan Message
	W chan<- Message
}

func MessageChansPair() (a, b MessageChans) {
	aToB := make(chan Message)
	bToA := make(chan Message)
	a = MessageChans{R: bToA, W: aToB}
	b = MessageChans{R: aToB, W: bToA}
	return
}

func Manage(lines lineio.LineChans) (msgs MessageChans) {
	msgs, myMsgs := MessageChansPair()
	go func() {
		defer close(myMsgs.W)
		for line := range lines.R {
			myMsgs.W <- ParseMessage(line)
		}
	}()
	go func() {
		defer close(lines.W)
		for msg := range myMsgs.R {
			lines.W <- msg.String()
		}
	}()
	return
}
