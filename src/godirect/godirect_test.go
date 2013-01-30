package godirect

import (
	"bytes"
	"io"
	. "launchpad.net/gocheck"
	"os"
	"strings"
	"syscall"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type GoDirectTests struct{}

var _ = Suite(&GoDirectTests{})

const _LOREM_IPSUM = `Lorem ipsum dolor sit amet, consectetur adipisicingelit, sed do eiusmod
tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,
quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
consequat. Duis aute irure dolor in reprehenderit in voluptate velit
esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat
cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est
laborum.`

type DummyReader struct {
	reader io.Reader
}

func (r *DummyReader) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		r.reader = strings.NewReader(_LOREM_IPSUM)
	}

	for {
		n, _ := r.reader.Read(p)

		if n == len(p) {
			break
		}

		r.reader = strings.NewReader(_LOREM_IPSUM)

		p = p[n:]
	}

	return len(p), nil
}

func NewDummyData(size int) []byte {
	b := make([]byte, size)
	new(DummyReader).Read(b)
	return b

}

func createTmpFile(c *C, fname string, data []byte) {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		c.Fatal("Could not write temp file: ", err)
	}

	_, err = f.Write(data)
	if err != nil {
		c.Fatal("Could not write temp file: ", err)
	}

	defer f.Close()
}

func (s *GoDirectTests) TestSimpleRead(c *C) {
	fname := "/var/tmp/test.tmp"
	var fsize int = 4096
	data := NewDummyData(fsize)
	createTmpFile(c, fname, data)
	defer os.Remove(fname)

	f, err := os.OpenFile(fname, os.O_RDONLY|syscall.O_DIRECT, 0)
	if err != nil {
		c.Fatal("Could not open file for direct io read: ", err)
	}
	defer f.Close()

	r := NewReader(f)
	buff := make([]byte, fsize)
	_, err = r.Read(buff)
	if err != nil {
		c.Fatal("Could not read file: ", err)
	}

	if !bytes.Equal(buff, data) {
		c.Errorf("Expected '%s' got '%s'", data, buff)
	}
}

func (s *GoDirectTests) TestContinousRead(c *C) {
	fname := "/var/tmp/test.tmp"
	fsize := 4096
	nreads := 4
	chunk := fsize / nreads
	data := NewDummyData(fsize)
	createTmpFile(c, fname, data)
	defer os.Remove(fname)

	f, err := os.OpenFile(fname, os.O_RDONLY|syscall.O_DIRECT, 0)
	if err != nil {
		c.Fatal("Could not open file for direct io read: ", err)
	}
	defer f.Close()

	r := NewReader(f)
	buff := make([]byte, fsize)

	for i := 0; i < nreads; i++ {
		_, err = r.Read(buff[i*chunk : (i+1)*chunk])
		if err != nil {
			c.Fatal("Could not read file: ", err)
		}
	}

	if !bytes.Equal(buff, data) {
		c.Errorf("Expected '%s' got '%s'", data, buff)
	}
}
