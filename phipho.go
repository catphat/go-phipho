package phipho

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	sys "golang.org/x/sys/unix"
)

// conventions:
// - If a method mutates the state of an existing struct, then it should be a pointer extension
//   TODO - get golang naming convention for "pointer extension"

// named pipe name
type name string

func (n *name) string() string {
	return fmt.Sprintf("%v", n)
}

func (n *name) getAbsPath() (p string, err error) {
	p, err = filepath.Abs(n.string())
	if err != nil {
		err = errors.Wrapf(err, "could not get absolute path of %v", n)
	}
	return p, err
}

func (n *name) getParentDir() (pd string, err error) {
	p, err := n.getAbsPath()
	if err != nil {
		return p, errors.Wrap(err, "could not get path")
	}

	pd, _ = filepath.Split(p)
	return p, nil
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
	err := np.n.initPipe()
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

func (n *name) initPipe() (err error) {
	err = sys.Mkfifo(n.string(), 0600)
	if err != nil {
		return errors.Wrap(err, "could not create pipe")
	}
	return nil
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
	f, err := n.getPipeRO()
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

func (p *np) writeln(s string) error {
	f, err := p.n.getPipeWO()
	if err != nil {
		return errors.Wrap(err, "could not get pipe for write only")
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%v\n", s))
	return err
}

func (n *name) getPipeRO() (f *os.File, err error) {
	f, err = n.getPipe(sys.O_NONBLOCK | sys.O_RDONLY | sys.O_EXCL)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return f, nil
}

// If O_NONBLOCK is set,  an open() for writing only will return an error
// if no process currently has the file open for reading.
// http://pubs.opengroup.org/onlinepubs/7908799/xsh/open.html
func (n *name) getPipeWO() (f *os.File, err error) {
	f, err = n.getPipe(syscall.O_APPEND | syscall.O_WRONLY)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return f, nil
}

func (n *name) getPipe(fileflags int) (file *os.File, err error) {
	fc := make(chan *os.File, 1)
	errc := make(chan error, 1)
	defer close(fc)
	defer close(errc)

	go func() {
		fp, err := n.getAbsPath()
		if err != nil {
			errc <- errors.Wrap(err, "could not get pipe absolute path")
			return
		}
		if file, err = os.OpenFile(fp, fileflags, os.ModeNamedPipe); os.IsNotExist(err) {
			errc <- errors.Wrap(err, "Named pipe does not exist")
			return
		} else if os.IsPermission(err) {
			errc <- errors.Wrapf(err, "Insufficient permissions to read named pipe '%s'", n)
			return
		} else if err != nil {
			errc <- errors.Wrapf(err, "Error while opening named pipe '%s'", n)
			return
		}
		fc <- file
	}()

	select {
	case err := <-errc:
		return nil, err
	case <-time.After(time.Millisecond * 100):
		return nil, errors.New("timeout while getting existing fifo file")
	case <-fc:
		return file, nil
	}
}
