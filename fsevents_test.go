package phipho

import (
	"fmt"
	"testing"
)

func TestEventHandlers(t *testing.T) {
	fse, err := newFsEvents("fsevents_test", "./fifo/")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan bool, 1)
	h := fse.fseHandlerFunc(func(e *fsEvent) {
		fmt.Printf("fsevent = %v\n", e)
		//done <- true
	})
	fse.opHandler(opWRITE, h)
	<-done
}

//type handlerFunc func(fsEvent)

//func (f handlerFunc) handle(e fsEvent) {
//	f(e)
//}
