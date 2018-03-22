package phipho

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func testNewFifo(t *testing.T, f *Fifo, name string) {
	err := f.New(name)
	if err != nil {
		t.Errorf("NewFifo(name) raised an error: %v", err)
	}
}

func testGetExistingFifoFile_willTimeout(t *testing.T, f *Fifo) {
	_, err := getExistingFifoFile(f.Name, syscall.O_RDONLY)
	expected := "timeout while getting existing fifo file"
	if err.Error() != expected {
		t.Errorf("GetExistingFile = %v, expected = %v", err, expected)
	}
}

func testGetExistingFifoFile(t *testing.T, f *Fifo) {
	_, err := getExistingFifoFile(f.Name, syscall.O_NONBLOCK|syscall.O_RDONLY)
	if err != nil {
		t.Error(err)
	}
}

func testFifoEvents(t *testing.T) {
	f := &Fifo{}
	err := f.New("testevents")
	if err != nil {
		t.Error(errors.Wrap(err, "could not make new Fifo"))
	}
	_, err = f.getExistingFifoFileRO()
	if err != nil {
		t.Error(errors.Wrap(err, "could not get existing fifo for RO"))
	}

	expectedEvents := []string{
		"notify.Write",
		"notify.Write",
		"notify.Remove",
	}

	expectedEventCount := (len(expectedEvents))

	aec := make(chan string, expectedEventCount+1) // +1 for notify.Stop()
	go func() {
		for ae := range f.Event {
			aec <- ae.Type
		}
		close(aec)
	}()

	for i := 0; i < expectedEventCount; i++ {
		var err error
		fmt.Printf("%v | %v\n", i, expectedEvents[i])
		switch ee := expectedEvents[i]; ee {
		case "notify.Write":
			err = f.sendStringMsg("test-write-msg")
		case "notify.Remove":
			err = f.Destroy()
			fmt.Println("executed destroy")
			if err != nil {
				t.Fatal(err)
			}
		case "notify.Create":
			// do nothing

		default:
			err = errors.New("unexpected default event case")
		}
		if err != nil {
			t.Fatal(errors.Wrap(err, "notify event loop error"))
		}
		//time.Sleep(time.Millisecond * 1)
	}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * 3)
		timeout <- true
	}()

	actualEvents := []string{}
	for ae := range aec {
		actualEvents = append(actualEvents, ae)
	}

	fmt.Println("here")
	fmt.Printf("actualEventcount %v\n", len(actualEvents))
	for i := 0; i < expectedEventCount; i++ {
		fmt.Printf("eventCount: %v\n", i)
		ee := expectedEvents[i]
		ae := actualEvents[i]
		fmt.Printf("#%v - ee %v | ae %v\n", i, ee, ae)

		if ae != ee {
			t.Fatalf("FifoEvent = %v, expected = %v", ae, ee)
		}

	}

}

func TestNewFifo(t *testing.T) {
	//f := &Fifo{}
	//defer f.Destroy()
	//	testNewFifo(t, f, "test")
	testFifoEvents(t)

}

//func TestGetExisting(t *testing.T) {
//	f := &Fifo{}
//	defer f.Destroy()
//	testNewFifo(t, f, "test")
//	testGetExistingFifoFile_willTimeout(t, f)
//	testGetExistingFifoFile(t, f)
//}
