package phipho

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	sys "golang.org/x/sys/unix"
)

func (n *name) makePipe() error {
	err := sys.Mkfifo(n.string(), 0600)
	if err != nil {
		return errors.Wrap(err, "could not create pipe")
	}
	return nil
}

func (n *name) deletePipe() error {
	return os.Remove(n.string())
}

func (n *name) getAbsPath() (p string, err error) {
	p, err = filepath.Abs(n.string())
	if err != nil {
		err = errors.Wrapf(err, "could not get absolute path of %v", n)
	}
	return p, err
}

// returns absolute parent directory
func (n *name) getParentDir() (pd string, err error) {
	p, err := n.getAbsPath()
	if err != nil {
		return p, errors.Wrap(err, "could not get path")
	}

	pd = filepath.Dir(p)
	return pd, nil
}

func (n *name) getPipeRO(nonBlock bool) (f *os.File, err error) {
	f, err = n.getPipe(sys.O_RDONLY|sys.O_EXCL, nonBlock)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return f, nil
}

func (n *name) getPipeWO(nonBlock bool) (f *os.File, err error) {
	f, err = n.getPipe(sys.O_APPEND|sys.O_WRONLY, nonBlock)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return f, nil
}

func (p *np) writeln(s string, nonBlock bool) error {
	f, err := p.n.getPipeWO(true)
	if err != nil {
		return errors.Wrap(err, "could not get pipe for write only")
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%v\n", s))
	return err
}

// If O_NONBLOCK is set,  an open() for writing only will return an error
// if no process currently has the file open for reading.
// http://pubs.opengroup.org/onlinepubs/7908799/xsh/open.html
func (n *name) getPipe(fileflags int, nonBlock bool) (file *os.File, err error) {
	fc := make(chan *os.File, 1)
	errc := make(chan error, 1)
	defer close(fc)
	defer close(errc)

	if nonBlock {
		fileflags = fileflags | syscall.O_NONBLOCK
	}

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

	if nonBlock {
		select {
		case err := <-errc:
			return nil, err
		case <-fc:
			return file, nil
		}
	} else {
		select {
		case err := <-errc:
			return nil, err
		case <-time.After(time.Millisecond * 100):
			return nil, errors.New("timeout while getting existing fifo file")
		case <-fc:
			return file, nil
		}
	}

}
