package phipho

import (
	"fmt"
	"os"
	"testing"
)

func TestFifo(t *testing.T) {

	np, err := newNp(
		Name("./fifo/testfifo"),
		//WriteEventHandler(nil),
		//ErrorEventHandler(nil),
	)

	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./fifo/testfifo")

	//	out, err := np.read()
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	// fmt.Println("reading")

	err = np.writeln("moi kaikki")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("wrote hello")

	//for s := range out {
	//	fmt.Println(s)
	//}
}

//func testNewFifo(t *testing.T, f *Fifo, name string) {
//	err := f.New(name)
//	if err != nil {
//		t.Errorf("NewFifo(name) raised an error: %v", err)
//	}
//}
//
//func testGetExistingFifoFile_willTimeout(t *testing.T, f *Fifo) {
//	_, err := getExistingFifoFile(f.Name, syscall.O_RDONLY)
//	expected := "timeout while getting existing fifo file"
//	if err.Error() != expected {
//		t.Errorf("GetExistingFile = %v, expected = %v", err, expected)
//	}
//}
//
//func testGetExistingFifoFile(t *testing.T, f *Fifo) {
//	_, err := getExistingFifoFile(f.Name, syscall.O_NONBLOCK|syscall.O_RDONLY)
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func testFifoEvents(t *testing.T) {
//	f := &Fifo{}
//	err := f.New("testevents")
//	if err != nil {
//		t.Fatal(errors.Wrap(err, "could not make new Fifo"))
//	}
//	file, err := f.getExistingFifoFileRO()
//	if err != nil {
//		t.Fatal(errors.Wrap(err, "could not get existing fifo for RO"))
//	}
//	defer file.Close()
//	defer f.Destroy()
//	expectedEvents := []string{
//		"WRITE",
//		"WRITE",
//		"REMOVE",
//	}
//
//	eec := (len(expectedEvents))
//
//	sendEvents := func(c int) {
//		for i := 0; i < c; i++ {
//			var err error
//			fmt.Printf("%v | %v\n", i, expectedEvents[i])
//			switch ee := expectedEvents[i]; ee {
//			case "WRITE":
//				f.Writeln(fmt.Sprintf("writing e: %v", i))
//			case "REMOVE":
//				err = f.Destroy() //todo this needs to be queued
//				fmt.Println("executed destroy")
//				if err != nil {
//					t.Fatal(err)
//				}
//			default:
//				err = fmt.Errorf("unexpected default event case: %v", ee)
//			}
//			if err != nil {
//				t.Fatal(errors.Wrap(err, "notify event loop error"))
//			}
//			//time.Sleep(time.Millisecond * 1)
//		}
//	}
//
//	aec := make(chan FifoEvent, 1)
//	watching := make(chan bool, 1)
//	go func() {
//		watching <- true
//		for ae := range f.Event {
//			aec <- ae
//		}
//		watching <- false
//	}()
//
//	actualEvents := []string{}
//	isWatching := true
//	sentEvents := false
//	for isWatching {
//		select {
//		case err := <-f.Error:
//			t.Fatal(err)
//		case ae := <-aec:
//			actualEvents = append(actualEvents, ae.Type)
//		case <-time.After(time.Second * 10):
//			t.Fatal("timeout")
//		case w := <-watching:
//			if w && !sentEvents {
//				sendEvents(eec)
//				sentEvents = true
//			}
//			isWatching = w
//		}
//	}
//
//	fmt.Printf("actualEventcount %v\n", len(actualEvents))
//	for i := 0; i < eec; i++ {
//		fmt.Printf("eventCount: %v\n", i)
//		ee := expectedEvents[i]
//		ae := actualEvents[i]
//		fmt.Printf("#%v - ee %v | ae %v\n", i, ee, ae)
//
//		if ae != ee {
//			t.Fatalf("FifoEvent = %v, expected = %v", ae, ee)
//		}
//
//	}
//
//}
//
//func TestNewFifo(t *testing.T) {
//	//f := &Fifo{}
//	//defer f.Destroy()
//	//	testNewFifo(t, f, "test")
//	testFifoEvents(t)
//
//}

//func TestGetExisting(t *testing.T) {
//	f := &Fifo{}
//	defer f.Destroy()
//	testNewFifo(t, f, "test")
//	testGetExistingFifoFile_willTimeout(t, f)
//	testGetExistingFifoFile(t, f)
//}
