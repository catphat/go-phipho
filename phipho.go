package phipho

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	sys "golang.org/x/sys/unix"
)

type Fi interface {
	In(b []byte) (n int, err error)
}

type Fo interface {
	Out(<-chan []byte) (err error)
}

// named pipe
type np struct {
	name      string // mkfifo file path
	events    <-chan string
	eventsErr <-chan error
}

func newNp(npPath string) (*np, error) {
	np := &np{}
	err := np.initNp("./fifo/testfifo")
	return np, err
}

func (p *np) initNp(npPath string) (err error) {
	p.name = npPath
	err = p.initPipe()
	if err != nil {
		return errors.Wrap(err, "could not init pipe")
	}
	p.events, p.eventsErr, err = p.initEvents()
	return err
}

func (p *np) initPipe() error {
	err := sys.Mkfifo(p.name, 0600)
	if err != nil {
		return errors.Wrap(err, "could not create pipe")
	}
	return nil
}

func (p *np) initEvents() (ec <-chan string, erc <-chan error, err error) {
	events := make(chan string, 1)
	eventsErr := make(chan error, 1)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return events, eventsErr, errors.Wrap(err, "could not create new watcher")
	}

	go func() {
		done := false
		for !done {
			select {
			case e := <-watcher.Events:
				eAbs, err := filepath.Abs(e.Name)
				if err != nil {
					eventsErr <- errors.Wrap(err, "fs notify event error")
				}

				pAbs, err := filepath.Abs(p.name)
				if err != nil {
					eventsErr <- errors.Wrap(err, "could get named pipe absolute path")
				}

				if eAbs != pAbs {
					continue
				}
				fmt.Println(e)
				events <- e.Op.String()
			case err := <-watcher.Errors:
				eventsErr <- errors.Wrap(err, "fs notify error")
			}
		}
	}()

	path, _ := filepath.Split(p.name)
	err = watcher.Add(path)
	if err != nil {
		err = errors.Wrapf(err, "could not add dir %v to watcher", path)
	}

	return events, eventsErr, err
}

func (p *np) writeln(s string) error {
	f, err := p.getPipeWO()
	if err != nil {
		return errors.Wrap(err, "could not get pipe for write only")
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%v\n", s))
	return err
}

func (p *np) read() (out <-chan string, err error) {
	oc := make(chan string, 1)
	f, err := p.getPipeRO()
	if err != nil {
		return nil, errors.Wrap(err, "could not get pipe for read only")
	}
	go func() {
		for {
			e := <-p.events
			switch e {
			case "WRITE":
				b := new(bytes.Buffer)
				b.ReadFrom(f)
				oc <- b.String()
			}
		}
	}()

	return oc, nil
}

func (p *np) getPipeRO() (f *os.File, err error) {
	f, err = p.getPipe(sys.O_NONBLOCK | sys.O_RDONLY | sys.O_EXCL)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return f, nil
}

// If O_NONBLOCK is set,  an open() for writing only will return an error
// if no process currently has the file open for reading.
// http://pubs.opengroup.org/onlinepubs/7908799/xsh/open.html
func (p *np) getPipeWO() (f *os.File, err error) {
	f, err = p.getPipe(syscall.O_APPEND | syscall.O_WRONLY)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return f, nil
}

func (p *np) getPipe(fileflags int) (file *os.File, err error) {
	fc := make(chan *os.File, 1)
	errc := make(chan error, 1)
	defer close(fc)
	defer close(errc)

	go func() {
		if file, err = os.OpenFile(p.name, fileflags, os.ModeNamedPipe); os.IsNotExist(err) {
			errc <- errors.Wrap(err, "Named pipe does not exist")
			return
		} else if os.IsPermission(err) {
			errc <- errors.Wrapf(err, "Insufficient permissions to read named pipe '%s'", p.name)
			return
		} else if err != nil {
			errc <- errors.Wrapf(err, "Error while opening named pipe '%s'", p.name)
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
