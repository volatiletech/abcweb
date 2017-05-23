package abcsessions

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/djherbis/times"
	uuid "github.com/satori/go.uuid"
)

var testpath string

func TestMain(m *testing.M) {
	var err error
	testpath, err = ioutil.TempDir("", "disksesstest")
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

	d, err := NewDiskStorer(filepath.Join(testpath, "a"), time.Hour*11, time.Hour*12)
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

	d, err := NewDiskStorer(filepath.Join(testpath, "b"), 0, 0)
	if err != nil {
		t.Error(err)
	}

	list, err := d.All()
	if err != nil {
		t.Error("expected no error on empty list")
	}
	if len(list) > 0 {
		t.Error("Expected len 0")
	}

	testid1 := uuid.NewV4().String()
	testid2 := uuid.NewV4().String()

	d.Set(testid1, "hello")
	d.Set(testid2, "friend")

	list, err = d.All()
	if err != nil {
		t.Error(err)
	}
	if len(list) != 2 {
		t.Errorf("Expected len 2, got %d", len(list))
	}
	if (list[0] != testid1 && list[0] != testid2) || list[0] == list[1] {
		t.Errorf("Expected list[0] to be %q or %q, got %q", "yo", "hi", list[0])
	}
	if (list[1] != testid2 && list[1] != testid1) || list[1] == list[0] {
		t.Errorf("Expected list[1] to be %q or %q, got %q", "hi", "yo", list[1])
	}
}

func TestDiskStorerGet(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(filepath.Join(testpath, "c"), 0, 0)
	if err != nil {
		t.Error(err)
	}

	val, err := d.Get("lol")
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	testid1 := uuid.NewV4().String()

	d.Set(testid1, "hello")

	val, err = d.Get(testid1)
	if err != nil {
		t.Error(err)
	}
	if val != "hello" {
		t.Errorf("Expected %q, got %s", "hello", val)
	}
}

func TestDiskStorerSet(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(filepath.Join(testpath, "d"), 0, 0)
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

	testid1 := uuid.NewV4().String()
	testid2 := uuid.NewV4().String()

	d.Set(testid1, "hello")
	d.Set(testid1, "whatsup")
	d.Set(testid2, "friend")

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected len 2, got %d", len(files))
	}

	val, err := d.Get(testid1)
	if err != nil {
		t.Error(err)
	}
	if val != "whatsup" {
		t.Errorf("Expected %q, got %s", "whatsup", val)
	}

	val, err = d.Get(testid2)
	if err != nil {
		t.Error(err)
	}
	if val != "friend" {
		t.Errorf("Expected %q, got %s", "friend", val)
	}
}

func TestDiskStorerDel(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(filepath.Join(testpath, "e"), 0, 0)
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

	testid1 := uuid.NewV4().String()
	testid2 := uuid.NewV4().String()

	d.Set(testid1, "hello")
	d.Set(testid1, "whatsup")
	d.Set(testid2, "friend")

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Errorf("Expected len 2, got %d", len(files))
	}

	err = d.Del(testid1)
	if err != nil {
		t.Error(err)
	}

	_, err = d.Get(testid1)
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
	d, err := NewDiskStorer(filepath.Join(testpath, "f"), time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	tm := diskTestTimer{}
	ch := make(chan time.Time)
	timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
		return tm, ch
	}

	testid1 := uuid.NewV4().String()
	testid2 := uuid.NewV4().String()

	//test1
	err = d.Set(testid1, testid1)
	if err != nil {
		t.Error(err)
	}
	//test2
	err = d.Set(testid2, testid2)
	if err != nil {
		t.Error(err)
	}

	// Change the mod time of test2 file to yesterday so we can test it gets deleted
	os.Chtimes(filepath.Join(d.folderPath, testid2),
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
	if files[0].Name() == testid2 {
		t.Errorf("expected test2 file to be deleted, but is present")
	}
}

func TestDiskStorerResetExpiry(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer(filepath.Join(testpath, "g"), 0, 0)
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

	testid1 := uuid.NewV4().String()

	err = d.Set(testid1, "val")
	if err != nil {
		t.Error(err)
	}

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected len 1, got %d", len(files))
	}

	ts, err := times.Stat(filepath.Join(d.folderPath, testid1))
	if err != nil {
		t.Error(err)
	}
	oldExpires := ts.AccessTime()

	time.Sleep(time.Nanosecond * 1)

	err = d.ResetExpiry(testid1)
	if err != nil {
		t.Error(err)
	}

	files, err = ioutil.ReadDir(d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected len 1, got %d", len(files))
	}

	ts, err = times.Stat(filepath.Join(d.folderPath, testid1))
	if err != nil {
		t.Error(err)
	}
	newExpires := ts.AccessTime()

	if !newExpires.After(oldExpires) || newExpires == oldExpires {
		t.Errorf("Expected newexpires to be newer than old expires, got: %#v, %#v", oldExpires, newExpires)
	}
}
