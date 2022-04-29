package app

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/c3llus/cdk/app/http"
)

type HTTPApp struct {
	*BaseApp
	service *http.Service
}

// New HTTP "APP"
func NewHTTP() (*HTTPApp, error) {

	var (
		err error
		ba  *BaseApp
	)

	ba, err = newBaseApp()
	if err != nil {
		return nil, err
	}

	return &HTTPApp{
		BaseApp: ba,
	}, nil
}

// Register HTTP Handlers
func (a *HTTPApp) RegisterServer(handler http.ServerHandler) error {

	if a.service != nil {
		return errors.New("HTTP service already registered")
	}

	a.service = http.New()
	a.service.RegisterService(handler)

	return nil
}

// Run application
// run will run all service in application
func (a *HTTPApp) Run() error {

	// creates http listener
	lis, err := net.Listen(network, port)
	if err != nil {
		return err
	}

	if err := a.service.Serve(lis); err != nil {
		// Error starting or closing listener:
		return err
	}

	// https://gobyexample.com/signals

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	fmt.Println("Service Started!")

	waitCtrlC(sigs, done)
	<-done

	return nil

}

func Listen(port string) (net.Listener, error) {
	return net.Listen("tcp4", port)
}

func waitCtrlC(sigs <-chan os.Signal, pressed chan bool) {
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		pressed <- true
	}()
}
