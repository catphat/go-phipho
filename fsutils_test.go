package phipho

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func check(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got = '%s' want = '%s'", got, want)
	}
}

func TestMakeAndDeletePipe(t *testing.T) {
	n := name("testpipe")

	t.Run("can create mkfifo pipe file", func(t *testing.T) {
		err := n.makePipe()
		if err != nil {
			t.Errorf("makePipe error = '%s'", err)
		}
	})

	t.Run("existing mkfifo pipe has correct permissions", func(t *testing.T) {
		info, err := os.Stat(n.string())
		if err != nil {
			t.Errorf("os.Stat error = '%s'", err)
		}

		got := info.Mode().Perm()
		want := os.FileMode(0600)
		if got != want {
			t.Errorf("incorrect permissions, got = '%v' want = '%v'", got, want)
		}
	})

	t.Run("can delete pipe", func(t *testing.T) {
		err := n.deletePipe()

		if err != nil {
			t.Errorf("delete pipe error = '%s'", err)
		}
	})
}

func TestPathInfo(t *testing.T) {
	parentDir := "/var/tmp/test-fsutils"
	pipeName := "test-fsutils-pipe"
	n := name(parentDir + "/" + pipeName)

	t.Run("has correct absolute path", func(t *testing.T) {
		got, err := n.getAbsPath()
		if err != nil {
			t.Errorf("could not get absolute path, '%s'", err)
		}

		want := parentDir + "/" + pipeName
		check(t, got, want)
	})

	t.Run("can get parent dir", func(t *testing.T) {
		got, err := n.getParentDir()
		if err != nil {
			t.Errorf("getParentDir error = '%s'", err)
		}

		want, err := filepath.Abs(parentDir)
		if err != nil {
			t.Errorf("could not get absolute path, '%s'", err)
		}
		check(t, got, want)
	})

}

func TestPipeRW(t *testing.T) {
	n := name("test-fsutitls-piperw")

	err := n.makePipe()
	if err != nil {
		t.Errorf("makePipe error = '%s'", err)
	}

	defer n.deletePipe()

	t.Run("cannot write to READ ONLY file", func(t *testing.T) {

		p, err := n.getPipeRO(true)
		if err != nil {
			t.Errorf("getPipeRO error = '%s'", err)
		}

		defer p.Close()

		fp, err := n.getAbsPath()
		if err != nil {
			t.Error(err)
		}

		_, err = p.Write([]byte("test"))
		got := err.Error()
		want := fmt.Sprintf("write %s: bad file descriptor", fp)

		check(t, got, want)
	})

	t.Run("can write to WRITE ONLY file", func(t *testing.T) {
		rp, err := n.getPipeRO(true) //have to have pipe open for reading or writing will fail with nonblock set
		if err != nil {
			t.Errorf("getPipeRO error = '%s'", err)
		}

		p, err := n.getPipeWO(false)
		if err != nil {
			t.Errorf("getPipeWO error = '%s'", err)
		}

		defer rp.Close()
		defer p.Close()

		_, err = p.Write([]byte("test"))
		if err != nil {
			t.Errorf("could not write to pipe, error = '%s'", err)
		}

	})

}
