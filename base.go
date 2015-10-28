package admiralproto

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// ExampleProtocolConnectTimeout is the timeout for the inital server connection
var ExampleProtocolConnectTimeout = time.Second * 10

// ExampleProtocol is an example structure that a server would implement
type ExampleProtocol struct {
	networkConnection net.Conn
	shutdownChan      chan struct{}
	readNetworkPipe   *io.PipeReader
	writeNetworkPipe  *io.PipeWriter
	readUserPipe      *io.PipeReader
	writeUserPipe     *io.PipeWriter

	closeConnection sync.Once
}

// NewExampleProtocol returns an instance of an example protocol
func NewExampleProtocol(networkProtocol, networkAddress string) *ExampleProtocol {
	var err error
	ep := &ExampleProtocol{shutdownChan: make(chan struct{})}

	if ep.networkConnection, err = net.DialTimeout(networkProtocol, networkAddress, ExampleProtocolConnectTimeout); err != nil {
		log.Println("Unable to connect to remote server", err)
		return nil
	}

	ep.readNetworkPipe, ep.writeNetworkPipe = io.Pipe()
	ep.readUserPipe, ep.writeUserPipe = io.Pipe()

	go ep.networkWriter()
	go ep.networkReader()

	return ep

}

// Close shuts down the protocol consumer
func (ep *ExampleProtocol) Close() {
	ep.closeConnection.Do(func() { close(ep.shutdownChan) })
	ep.networkConnection.Close()
	ep.readUserPipe.Close()
	ep.readNetworkPipe.Close()
}

func (ep *ExampleProtocol) networkWriter() {
	for {
		select {
		case <-ep.shutdownChan:
			log.Println("Network Writer Shutting Down")
			ep.readNetworkPipe.Close()
			return
		default:
			if _, err := io.Copy(ep.networkConnection, ep.readNetworkPipe); err != nil && err != io.EOF {
				log.Println("Network Write Error", err)
				ep.readNetworkPipe.CloseWithError(err)
				ep.Close()
			}
		}
	}
}

func (ep *ExampleProtocol) networkReader() {
	for {
		select {
		case <-ep.shutdownChan:
			log.Println("Shutting down network reader")
			ep.writeNetworkPipe.Close()
			return
		default:
			if _, err := io.Copy(ep.writeUserPipe, ep.networkConnection); err != nil && err != io.EOF {
				log.Println("Error writing to internal pipe from network", err)
				ep.writeUserPipe.CloseWithError(err)
				ep.Close()
			}
		}
	}
}

// Read pulls data into from the internal pipe (which is copied from the socket)
func (ep *ExampleProtocol) Read(data []byte) (n int, err error) {
	return ep.readUserPipe.Read(data)

}

// Write pushes data into the internal pipe (which is then copied to the socket)
func (ep *ExampleProtocol) Write(data []byte) (n int, err error) {
	return ep.writeNetworkPipe.Write(data)
}
