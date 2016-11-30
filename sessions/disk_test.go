package sessions

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestMain(m *testing.M) {
	// Use an in-memory filesystem for testing so we don't pollute the disk
	fs = afero.NewMemMapFs()

	retCode := m.Run()

	os.Exit(retCode)
}

func TestDiskStorerNew(t *testing.T) {
	t.Parallel()

	d, err := NewDiskStorer("path", time.Hour*11, time.Hour*12)
	if err != nil {
		t.Error(err)
	}

	if d.maxAge != time.Hour*11 {
		t.Errorf("expected max age to be %d", time.Hour*11)
	}
	if d.folderPath != "path" {
		t.Errorf("expected folder path to be %q", "path")
	}

	d.wg.Wait()
}

func TestDiskStorerNewDefault(t *testing.T) {
	t.Parallel()

	d, err := NewDefaultDiskStorer()
	if err != nil {
		t.Error(err)
	}

	if d.maxAge != time.Hour*24*7 {
		t.Error("expected max age to be a week")
	}

	d.wg.Wait()
}

func TestDiskStorerGet(t *testing.T) {
	t.Parallel()

	d, _ := NewDiskStorer(path.Join(os.TempDir(), "a"), 0, 0)

	val, err := d.Get("lol")
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	d.Put("hi", "hello")

	val, err = d.Get("hi")
	if err != nil {
		t.Error(err)
	}
	if val != "hello" {
		t.Errorf("Expected %q, got %s", "hello", val)
	}
}

func TestDiskStorerPut(t *testing.T) {
	t.Parallel()

	d, _ := NewDiskStorer(path.Join(os.TempDir(), "b"), 0, 0)

	files, err := afero.ReadDir(fs, d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 0 {
		t.Errorf("Expected len 0, got %d", len(files))
	}

	d.Put("hi", "hello")
	d.Put("hi", "whatsup")
	d.Put("yo", "friend")

	files, err = afero.ReadDir(fs, d.folderPath)
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

	d, _ := NewDiskStorer(path.Join(os.TempDir(), "c"), 0, 0)

	files, err := afero.ReadDir(fs, d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 0 {
		t.Errorf("Expected len 0, got %d", len(files))
	}

	d.Put("hi", "hello")
	d.Put("hi", "whatsup")
	d.Put("yo", "friend")

	files, err = afero.ReadDir(fs, d.folderPath)
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

	files, err = afero.ReadDir(fs, d.folderPath)
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
	d, _ := NewDiskStorer(path.Join(os.TempDir(), "d"), time.Hour, time.Hour)

	tm := diskTestTimer{}
	ch := make(chan time.Time)
	timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
		return tm, ch
	}

	err := d.Put("testid1", "test1")
	if err != nil {
		t.Error(err)
	}
	err = d.Put("testid2", "test2")
	if err != nil {
		t.Error(err)
	}

	// Change the mod time of testid2 file to yesterday so we can test it gets deleted
	fs.Chtimes(path.Join(d.folderPath, "testid2"),
		time.Now().AddDate(0, 0, -1),
		time.Now().AddDate(0, 0, -1),
	)

	// Ensure there are currently 2 files, as expected
	files, err := afero.ReadDir(fs, d.folderPath)
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

	files, err = afero.ReadDir(fs, d.folderPath)
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		for _, f := range files {
			t.Log(f.Name())
		}
		t.Fatalf("Expected len 1, got %d: %#v", len(files), files)

	}
	if files[0].Name() != "testid1" {
		t.Errorf("expected testid2 to be deleted, but is present")
	}
}
