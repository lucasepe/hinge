package hinge

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type PluginSet map[string]Plugin

type ServeConfig struct {
	Plugins PluginSet
	Verbose bool
}

func Serve(opts *ServeConfig) {
	exitCode := -1

	defer func() {
		if exitCode >= 0 {
			os.Exit(exitCode)
		}
	}()

	logger := log.New(os.Stderr, "[plugin-server] ", log.LstdFlags)
	if !opts.Verbose {
		log.SetOutput(io.Discard)
		log.SetFlags(0)
	}

	lis, err := serverListener()
	if err != nil {
		return
	}
	defer func() {
		lis.Close()
	}()

	doneCh := make(chan struct{})
	var stdoutReader, stderrReader io.Reader
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}

	server := &RPCServer{
		Plugins: opts.Plugins,
		Stdout:  stdoutReader,
		Stderr:  stderrReader,
		DoneCh:  doneCh,
		logger:  logger,
	}

	if err := server.Init(); err != nil {
		logger.Println("protocol init ", "error ", err.Error())
		return
	}

	logger.Println("plugin address ", "network ",
		lis.Addr().Network(), "address ", lis.Addr().String())

	// Output the address and service name to stdout so that the client can
	// bring it up.
	fmt.Printf("%s|%s\n",
		lis.Addr().Network(),
		lis.Addr().String())
	os.Stdout.Sync()

	// Set our stdout, stderr to the stdio stream that clients can retrieve
	// using ClientConfig.SyncStdout/err.
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	// Accept connections and wait for completion
	go server.Serve(lis)

	ctx := context.Background()
	select {
	case <-ctx.Done():
		lis.Close()
		<-doneCh
	case <-doneCh:
	}
}

func serverListener() (net.Listener, error) {
	tf, err := os.CreateTemp("", "plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	// Wrap the listener in rmListener so that the Unix domain socket file
	// is removed on close.
	return &rmListener{
		Listener: l,
		Path:     path,
	}, nil
}

// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

func (l *rmListener) Close() error {
	// Close the listener itself
	if err := l.Listener.Close(); err != nil {
		return err
	}

	// Remove the file
	return os.Remove(l.Path)
}
