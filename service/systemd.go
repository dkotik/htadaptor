package service

import (
	"errors"
	"fmt"

	"github.com/coreos/go-systemd/activation"
)

// WithSystemDSocketActivationPort binds the server to a listener specified by systemd configuration. This will preserve TCP connections during service restarts.
//
//	// /lib/systemd/system/myapp.socket
//	[Socket]
//	ListenStream = 80
//	#ListenStream = 443
//	BindIPv6Only = both
//	Service      = myapp.service
//
//	[Install]
//	WantedBy = sockets.target
//
//	// myapp.service
//	[Unit]
//	Description = myapp
//	After       = network.target
//
//	[Service]
//	Type = simple
//
//	ExecStart = /bin/myapp
//	ExecStop  = /bin/kill $MAINPID
//	KillMode  = none
//
//	[Install]
//	WantedBy = multi-user.target
//
// See: https://bunrouter.uptrace.dev/guide/go-zero-downtime-restarts.html
//
// sudo systemctl start myapp.socket
// sudo systemctl status myapp.socket
// sudo systemctl restart myapp.socket
func WithFirstSystemDSocketActivationSocket() Option {
	return func(o *options) error {
		if o.Listener != nil {
			return errors.New("server address is already set")
		}
		listeners, err := activation.Listeners()
		if err != nil {
			return fmt.Errorf("could not access systemd network listeners: %w", err)
		}
		if len(listeners) == 0 {
			return errors.New("systemd service has no associated network listeners")
		}

		o.Listener = listeners[0]
		return nil
	}
}

// // TODO: below seems to produce a list of listeners by file name, probably not useful differentiation.
// func WithSystemDSocketActivationSocket(name string) Option {
// 	return func(o *options) error {
// 		if o.Listener != nil {
// 			return errors.New("server address is already set")
// 		}
// 		if name == "" {
// 			return errors.New("systemd socket name is required")
// 		}
// 		listeners, err := activation.ListenersWithNames()
// 		if err != nil {
// 			return fmt.Errorf("could not access systemd network listeners: %w", err)
// 		}
// 		if len(listeners) == 0 {
// 			return errors.New("systemd service has no associated network listeners")
// 		}
//
// 		ln, ok := listeners[name]
// 		if !ok {
// 			return fmt.Errorf("socket named %q is not associated with systemd service", name)
// 		}
//
// 		o.Listener = ln[0]
// 		return nil
// 	}
// }
