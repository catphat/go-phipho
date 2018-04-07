package phipho

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

// conventions:
// - If a method mutates the state of an existing struct, then it should be a pointer extension
//   TODO - get golang naming convention for "pointer extension"

// named pipe name
type name string

func (n *name) string() string {
	return fmt.Sprintf("%v", *n)
}

// named pipe
type np struct {
	n name // mkfifo filepath/name, if nil then a default value will be used (pid)
	e *fsEvents
	//weh *pipeWriteEventHandler //fs WRITE event handler, if nil then default handler will be used
	//eeh *pipeErrorEventHandler //fs error event handler, if nil then default handler will be used
}

func newNp(opts ...option) (*np, error) {
	np := &np{}
	np.Options(opts)
	err := np.n.makePipe()
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize pipe")
	}
	p, err := np.n.getParentDir()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get parent dir of %v", np.n)
	}
	e, err := newFsEvents(np.n.string(), p)
	if err != nil {
		return nil, errors.Wrap(err, "could create new File Systems Event watcher")
	}
	np.e = e
	return np, err
}

type option func(*np)

func (np *np) Options(opts []option) {
	for _, opt := range opts {
		np.Option(opt)
	}
}

func (np *np) Option(opts ...option) {
	for _, opt := range opts {
		opt(np)
	}
}

func Name(fileName string) option {
	return func(np *np) {
		np.n = name(fileName)
	}
}

type pipeFileReader interface {
	read(n *name, stop <-chan bool) (out <-chan string, err error)
}

func readFromFile(n *name, stop <-chan bool) (out <-chan string, err error) {
	oc := make(chan string, 1)
	f, err := n.getPipeRO(false)
	if err != nil {
		return nil, errors.Wrap(err, "could not get file for read only")
	}
	go func() {
		select {
		case <-stop:
			f.Close()
			close(oc)
			return
		default:
			b := new(bytes.Buffer)
			b.ReadFrom(f)
			oc <- b.String()
		}
	}()
	return oc, nil
}

//func (p *np) read() (out <-chan string, err error) {
//	oc := make(chan string, 1)
//	f, err := p.n.getPipeRO()
//	if err != nil {
//		return nil, errors.Wrap(err, "could not get pipe for read only")
//	}
//	go func() {
//		for {
//			e := <-p.e.c
//			switch e {
//			case "WRITE":
//				b := new(bytes.Buffer)
//				b.ReadFrom(f)
//				oc <- b.String()
//			}
//		}
//	}()
//
//	return oc, nil
//}
