package gauja

type TimeoutSettings struct {
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

func Connect(address string, timeoutSettings TimeoutSettings) (net.Conn, error) {
	return net.DialTimeout("tcp", address, timeoutSettings.ConnectTimeout)
}
