package phipho

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rjeczalik/notify"
)

// Fifo represents a named mkfifo pipe
type Fifo struct {
	Name  string
	Event chan FifoEvent
}

type FifoEvent struct {
	Type string
	Info string
}

func (f *Fifo) New(name string) (err error) {
	f.Name = name
	dir := path.Dir(".") //todo this should be an argument
	abs, err := filepath.Abs(path.Join(dir, name))
	if err != nil {
		return errors.Wrap(err, "could not get absolute path")
	}

	err = syscall.Mkfifo(f.Name, 0600)
	if err != nil {
		return errors.Wrap(err, "could not create mkfifo pipe")
	}

	ec := make(chan notify.EventInfo, 4)
	if err = notify.Watch(dir, ec, notify.All); err != nil {
		fmt.Println(err)
		return errors.Wrap(err, "could not watch fifo file.")
	}

	f.Event = make(chan FifoEvent, 4)

	go func() {
		for {
			ei := <-ec

			if ei.Path() != abs {
				continue
			}
			fmt.Println(ei)
			f.Event <- FifoEvent{
				Type: fmt.Sprintf("%s", ei.Event()),
				Info: ei.Path(),
			}
			if ei.Event() == notify.Remove {
				defer notify.Stop(ec)
				defer close(f.Event)
				fmt.Println("stopped ec....")
				break
			}
		}
	}()

	//	for {
	//		fmt.Println("waiting for events...")
	//		event := <-f.Event
	//		fmt.Println(event)
	//	}
	return nil
}

func (f *Fifo) getExistingFifoFileRO() (file *os.File, err error) {
	file, err = getExistingFifoFile(f.Name, syscall.O_NONBLOCK|syscall.O_RDONLY|syscall.O_EXCL)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return file, nil
}

func (f *Fifo) getExistingFifoFileWO() (file *os.File, err error) {
	// If O_NONBLOCK is set,  an open() for writing only will return an error
	// if no process currently has the file open for reading.
	file, err = getExistingFifoFile(f.Name, syscall.O_NONBLOCK|syscall.O_WRONLY)
	if err != nil {
		return nil, errors.Wrap(err, "could not get existing fifo file")
	}
	return file, nil
}

func getExistingFifoFile(name string, fileflags int) (file *os.File, err error) {
	fc := make(chan *os.File, 1)
	errc := make(chan error, 1)
	go func() {

		// http://pubs.opengroup.org/onlinepubs/007908799/xsh/open.html (O_NONBLOCK)
		// syscall.O_NONBLOCK
		if file, err = os.OpenFile(name, fileflags, os.ModeNamedPipe); os.IsNotExist(err) {
			errc <- errors.Wrap(err, "Named pipe does not exist")
			return
		} else if os.IsPermission(err) {
			errc <- errors.Wrap(err, fmt.Sprintf("Insufficient permissions to read named pipe '%s'", name))
			return
		} else if err != nil {
			errc <- errors.Wrap(err, fmt.Sprintf("Error while opening named pipe '%s'", name))
			return
		}
		fc <- file

	}()

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Millisecond * 100)
		timeout <- true
	}()

	select {
	case err := <-errc:
		return nil, err
	case <-timeout:
		return nil, errors.New("timeout while getting existing fifo file")
	case <-fc:
		return file, nil
	}
}

func (f *Fifo) write(b []byte) error {
	file, err := f.getExistingFifoFileWO()
	if err != nil {
		return errors.Wrap(err, "failed openening fifo file")
	}
	defer file.Close()
	_, err = file.Write(b)
	if err != nil {
		return errors.Wrap(err, "failed writing to fifo file")
	}
	return nil
}

// Append write with newline
func (f *Fifo) writeln(b []byte) error {
	b = append(b, byte(10))
	return f.write(b)
}

func (f *Fifo) sendStringMsg(msg string) error {
	return f.writeln([]byte(msg))
}

func (f *Fifo) sendIntMsg(msg int) error {
	strMsg := strconv.FormatInt(int64(msg), 10)
	return f.sendStringMsg(strMsg)
}

func (f *Fifo) Destroy() error {
	err := os.Remove(f.Name)
	if err != nil {
		return errors.Wrap(err, "could not remove file")
	}
	return nil

}
