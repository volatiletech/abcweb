package sessions

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

var testpath string

func TestMain(m *testing.M) {
	testpath = path.Join(os.TempDir(), "disksesstest")
	err := os.Mkdir(testpath, os.FileMode(int(0755)))
	if err != nil {
		panic(err)
	}

	retCode := m.Run()

	err = os.RemoveAll(testpath)
	if err != nil {
		panic(err)
	}

	os.Exit(retCode)
}

func TestDiskStorerNew(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(path.Join(testpath, "a"), time.Hour*11, time.Hour*12)
	if err != nil {
		t.Error(err)
	}

	if d.maxAge != time.Hour*11 {
		t.Errorf("expected max age to be %d", time.Hour*11)
	}
	if d.folderPath != testpath+"/a" {
		t.Errorf("expected folder path to be %q", testpath+"/a")
	}

	d.wg.Wait()
}

func TestDiskStorerAll(t *testing.T) {
	t.Parallel()

	t.Error("not implemented")
}

func TestDiskStorerGet(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(path.Join(testpath, "b"), 0, 0)
	if err != nil {
		t.Error(err)
	}

	val, err := d.Get("lol")
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	d.Set("hi", "hello")

	val, err = d.Get("hi")
	if err != nil {
		t.Error(err)
	}
	if val != "hello" {
		t.Errorf("Expected %q, got %s", "hello", val)
	}
}

func TestDiskStorerSet(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(path.Join(testpath, "c"), 0, 0)
	if err != nil {
		t.Error(err)
	}

	files, err := ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 0 {
		t.Errorf("Expected len 0, got %d", len(files))
	}

	d.Set("hi", "hello")
	d.Set("hi", "whatsup")
	d.Set("yo", "friend")

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected len 2, got %d", len(files))
	}

	val, err := d.Get("hi")
	if err != nil {
		t.Error(err)
	}
	if val != "whatsup" {
		t.Errorf("Expected %q, got %s", "whatsup", val)
	}

	val, err = d.Get("yo")
	if err != nil {
		t.Error(err)
	}
	if val != "friend" {
		t.Errorf("Expected %q, got %s", "friend", val)
	}
}

func TestDiskStorerDel(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(path.Join(testpath, "d"), 0, 0)
	if err != nil {
		t.Error(err)
	}

	files, err := ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 0 {
		t.Errorf("Expected len 0, got %d", len(files))
	}

	d.Set("hi", "hello")
	d.Set("hi", "whatsup")
	d.Set("yo", "friend")

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected len 2, got %d", len(files))
	}

	err = d.Del("hi")
	if err != nil {
		t.Error(err)
	}

	_, err = d.Get("hi")
	if err == nil {
		t.Errorf("Expected get hi to fail")
	}

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected len 1, got %d", len(files))
	}
}

// diskTestTimer is used in the timerTestHarness override so we can
// control sending signals to the sleep channel and trigger cleans manually
type diskTestTimer struct{}

func (diskTestTimer) Reset(time.Duration) bool {
	return true
}

func (diskTestTimer) Stop() bool {
	return true
}

func TestDiskStorerCleaner(t *testing.T) {
	d, err := NewDiskStorer(path.Join(testpath, "e"), time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	tm := diskTestTimer{}
	ch := make(chan time.Time)
	timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
		return tm, ch
	}

	err = d.Set("testid1", "test1")
	if err != nil {
		t.Error(err)
	}
	err = d.Set("testid2", "test2")
	if err != nil {
		t.Error(err)
	}

	// Change the mod time of testid2 file to yesterday so we can test it gets deleted
	os.Chtimes(path.Join(d.folderPath, "testid2"),
		time.Now().AddDate(0, 0, -1),
		time.Now().AddDate(0, 0, -1),
	)

	// Ensure there are currently 2 files, as expected
	files, err := ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected len 2, got %d", len(files))
	}

	// Start the cleaner go routine
	d.StartCleaner()

	// Signal the timer channel to execute the clean
	ch <- time.Time{}

	// Stop the cleaner, this will block until the cleaner has finished its operations
	d.StopCleaner()

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		for _, f := range files {
			t.Log(f.Name())
		}
		t.Errorf("Expected len 1, got %d: %#v", len(files), files)

	}
	if files[0].Name() != "testid1" {
		t.Errorf("expected testid2 to be deleted, but is present")
	}
}

func TestDiskStorerResetExpiry(t *testing.T) {
	t.Parallel()

	t.Error("not implemented")
}
