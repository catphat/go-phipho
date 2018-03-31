package phipho

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// file system events
type fsEvents struct {
	//	c          <-chan fsEvent // fs events channel
	//	ec         <-chan error   // fs event errors channel
	watchFile  string // target file
	watchDir   string // parent directory for watchFile, events will be generated by watching this directory
	opHandlers map[fsOp][]fsEventHandler
}

type fsOp string

const (
	opCREATE fsOp = "CREATE"
	opWRITE  fsOp = "WRITE"
	opREMOVE fsOp = "REMOVE"
	opRENAME fsOp = "RENAME"
	opCHMOD  fsOp = "CHMOD"
)

type fsEvent struct {
	name string // name of file, typically path/filename
	op   fsOp   // filesystem event operation
}

// n - name of file
// p - parent dir of file
func newFsEvents(n, p string) (*fsEvents, error) {
	fse := &fsEvents{
		//	c:  make(chan fsEvent, 1),
		//	ec: make(chan error, 1),
	}

	err := fse.init(n, p)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize events")
	}
	return fse, nil
}

func getFsOp(fsnOp fsnotify.Op) (fsOp, error) {
	var op fsOp
	switch fsnOp {
	case fsnotify.Create:
		op = opCREATE
	case fsnotify.Write:
		op = opWRITE
	case fsnotify.Remove:
		op = opREMOVE
	case fsnotify.Rename:
		op = opRENAME
	case fsnotify.Chmod:
		op = opCHMOD
	default:
		return op, errors.Errorf("could not get filesystem operation for fs notify type %v", fsnOp)
	}

	return op, nil
}

// fn - filename of mkfifo pipe
// p - parent directory of pipe's file
func (fse *fsEvents) init(fn string, p string) (err error) {
	ec := make(chan fsEvent, 1)
	erc := make(chan error, 1)
	done := make(chan bool, 1)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "could not create new watcher")
	}

	go func() {
		for {
			select {
			case <-done:
				fmt.Println("closing")
				watcher.Close()
				close(ec)
				close(erc)
				return

			case e := <-watcher.Events:
				op, err := getFsOp(e.Op)
				if err != nil {
					erc <- errors.Wrapf(err, "could not get op type for %v", e.Name)
					continue
				}

				fse := fsEvent{
					name: e.Name,
					op:   op,
				}
				fmt.Printf("event %v\n", fse)
				ec <- fse
				//eAbs, err := filepath.Abs(fse.Name)
				//if err != nil {
				//	erc <- errors.Wrap(err, "fs notify event error")
				//}

				//pAbs, err := n.getAbsPath()
				//if err != nil {
				//	erc <- errors.Wrap(err, "could get not named pipe path")
				//}

				//if eAbs != pAbs {
				//	continue
				//}
				//fmt.Println(fse)
				//ec <- fse.Op.String()

			case err := <-watcher.Errors:
				erc <- errors.Wrap(err, "fs notify error")
			}
		}

	}()

	eventHandlers := func(c <-chan fsEvent) {
		for e := range c {
			if _, ok := fse.opHandlers[e.op]; ok {
				for _, h := range fse.opHandlers[e.op] {
					go func(h fsEventHandler) {
						h.handle(&e)
					}(h)
				}
			}
		}
	}

	go func(ec <-chan fsEvent) {
		eventHandlers(ec)
	}(ec)

	go func(erc <-chan error) {
		fmt.Printf("ERROR: %v\n", <-erc)
	}(erc)

	fmt.Printf("adding %v to watcher\n", p)
	err = watcher.Add(p)
	if err != nil {
		return errors.Wrapf(err, "could not add dir %v to watcher", p)
	}

	fse.c = ec
	fse.ec = erc
	return nil
}

type fsEventHandler interface {
	handle(*fsEvent)
}

type fsEventHandlerFunc func(*fsEvent)

func (f fsEventHandlerFunc) handle(e *fsEvent) {
	f(e)
}

func (fse *fsEvents) fseHandlerFunc(f fsEventHandlerFunc) fsEventHandler {
	return f
}

func (fse *fsEvents) opHandler(op fsOp, h fsEventHandler) {
	if fse.opHandlers == nil {
		fse.opHandlers = make(map[fsOp][]fsEventHandler)
	}

	if _, ok := fse.opHandlers[op]; ok {
		fse.opHandlers[op] = append(fse.opHandlers[op], h)
	} else {
		fse.opHandlers[op] = []fsEventHandler{h}
	}
}

//func (fse *fsevents) initEventHandlers(weh pipeWriteEventHandler, eeh pipeErrorEventHandler) {
//
//	go func() {
//		select {
//		case e := <-fse.c:
//			switch e {
//			case "WRITE":
//				// handle writes
//				fmt.Printf("EVENT: WRITE = %v\n", e)
//				weh.handle()
//			}
//		case err := <-fse.ec:
//			// handle errors
//			fmt.Printf("EVENT: WRITE = %v\n", err)
//			eeh.handle(err)
//
//		}
//	}()
//}
