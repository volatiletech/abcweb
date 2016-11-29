package sessions

import (
	"os"
	"path"
	"sync"
	"time"

	"github.com/spf13/afero"
)

// DiskStorer is a session storer implementation for saving sessions
// to disk.
type DiskStorer struct {
	// Path to the session files folder
	folderPath string
	// How long sessions take to expire on disk
	maxAge time.Duration
	// Disk storage mutex
	mut sync.RWMutex
	// wg is used to manage the cleaner go routines
	wg sync.WaitGroup
}

// FS is a filesystem pointer. This is used in favor of os and ioutil directly
// so that we can point it to a mock filesystem in the tests to avoid polluting
// the disk when testing.
var FS afero.Fs = afero.NewOsFs()

// NewDefaultDiskStorer returns a DiskStorer object with default values.
// The default values are:
// FolderPath: system tmp dir + random folder
// maxAge: 1 week (clear session stored on server after 1 week)
// cleanInterval: 2 hours (delete sessions older than maxAge every 2 hours)
func NewDefaultDiskStorer() (*DiskStorer, error) {
	folderPath := path.Join(os.TempDir(), "UNIQUEIDHERE")
	return NewDiskStorer(folderPath, time.Hour*24*7, time.Hour*2)
}

// NewDiskStorer initializes and returns a new DiskStorer object.
// It takes the maxAge of how long each session should live on disk,
// and a cleanInterval duration which defines how often the clean
// task should check for maxAge expired sessions to be removed from disk.
// Persistent storage can be attained by setting maxAge and cleanInterval
// to zero.
func NewDiskStorer(folderPath string, maxAge, cleanInterval time.Duration) (*DiskStorer, error) {
	if (maxAge != 0 && cleanInterval == 0) || (cleanInterval != 0 && maxAge == 0) {
		panic("if max age or clean interval is set, the other must also be set")
	}

	d := &DiskStorer{
		folderPath: folderPath,
		maxAge:     maxAge,
	}

	// Create the storage folder if it does not exist
	_, err := FS.Stat(folderPath)
	if os.IsNotExist(err) {
		err := FS.Mkdir(folderPath, os.FileMode(int(0755)))
		if err != nil {
			return nil, err
		}
	}

	// If max age is set start the memory cleaner go routine
	if int(maxAge) != 0 {
		d.wg.Add(1)
		go d.cleaner(cleanInterval)
	}

	return d, nil
}

// Get returns the value string saved in the session pointed to by the
// session id key.
func (d *DiskStorer) Get(key string) (value string, err error) {
	filePath := path.Join(d.folderPath, key)

	d.mut.RLock()
	defer d.mut.RUnlock()

	_, err = FS.Stat(filePath)
	if os.IsNotExist(err) {
		return "", errNoSession{}
	} else if err != nil {
		return "", err
	}

	contents, err := afero.ReadFile(FS, filePath)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

// Put saves the value string to the session pointed to by the session id key.
func (d *DiskStorer) Put(key, value string) error {
	filePath := path.Join(d.folderPath, key)

	d.mut.Lock()
	defer d.mut.Unlock()

	return afero.WriteFile(FS, filePath, []byte(value), 0600)
}

// Del the session pointed to by the session id key and remove it.
func (d *DiskStorer) Del(key string) error {
	filePath := path.Join(d.folderPath, key)

	d.mut.Lock()
	defer d.mut.Unlock()

	_, err := FS.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return FS.Remove(filePath)
}

// diskSleepFunc is a test harness
var diskSleepFunc = func(sleep time.Duration) bool {
	time.Sleep(sleep)
	return true
}

func (d *DiskStorer) cleaner(loop time.Duration) {
	for {
		if ok := diskSleepFunc(loop); !ok {
			defer d.wg.Done()
			return
		}

		t := time.Now().UTC()

		d.mut.RLock()
		files, err := afero.ReadDir(FS, d.folderPath)
		d.mut.RUnlock()

		if err != nil {
			panic(err)
		}

		for _, file := range files {
			// File is expired
			if file.ModTime().UTC().Add(d.maxAge).Before(t) {
				filePath := path.Join(d.folderPath, file.Name())

				d.mut.Lock()
				_, err := FS.Stat(filePath)
				// If the file has been deleted manually from the server
				// in between the time we read the directory and now, it will
				// fail here with a ErrNotExist. If so, continue gracefully.
				if os.IsNotExist(err) {
					d.mut.Unlock()
					continue
				} else if err != nil {
					panic(err)
				}

				err = FS.Remove(filePath)
				d.mut.Unlock()
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
