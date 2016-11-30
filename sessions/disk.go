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
	// How often the disk should be polled for maxAge expired sessions
	cleanInterval time.Duration
	// Disk storage mutex
	mut sync.RWMutex
	// wg is used to manage the cleaner loop
	wg sync.WaitGroup
	// quit channel for exiting the cleaner loop
	quit chan struct{}
}

// fs is a filesystem pointer. This is used in favor of os and ioutil directly
// so that we can point it to a mock filesystem in the tests to avoid polluting
// the disk when testing.
var fs = afero.NewOsFs()

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
		folderPath:    folderPath,
		maxAge:        maxAge,
		cleanInterval: cleanInterval,
	}

	// Create the storage folder if it does not exist
	_, err := fs.Stat(folderPath)
	if os.IsNotExist(err) {
		err := fs.Mkdir(folderPath, os.FileMode(int(0755)))
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}

// Get returns the value string saved in the session pointed to by the
// session id key.
func (d *DiskStorer) Get(key string) (value string, err error) {
	filePath := path.Join(d.folderPath, key)

	d.mut.RLock()
	defer d.mut.RUnlock()

	_, err = fs.Stat(filePath)
	if os.IsNotExist(err) {
		return "", errNoSession{}
	} else if err != nil {
		return "", err
	}

	contents, err := afero.ReadFile(fs, filePath)
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

	return afero.WriteFile(fs, filePath, []byte(value), 0600)
}

// Del the session pointed to by the session id key and remove it.
func (d *DiskStorer) Del(key string) error {
	filePath := path.Join(d.folderPath, key)

	d.mut.Lock()
	defer d.mut.Unlock()

	_, err := fs.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return fs.Remove(filePath)
}

// StopCleaner stops the cleaner go routine
func (d *DiskStorer) StopCleaner() {
	close(d.quit)
	d.wg.Wait()
}

// StartCleaner starts the disk session cleaner go routine. This go routine
// will delete expired disk sessions on the cleanInterval interval.
func (d *DiskStorer) StartCleaner() {
	if d.maxAge == 0 || d.cleanInterval == 0 {
		panic("both max age and clean interval must be set to non-zero")
	}

	// init quit chan
	d.quit = make(chan struct{})

	d.wg.Add(1)

	// Start the cleaner infinite loop go routine.
	// StopCleaner() can be used to kill this go routine.
	go d.cleanerLoop()
}

// cleanerLoop executes the Clean() method every time cleanInterval elapses.
// StopCleaner() can be used to kill this go routine loop.
func (d *DiskStorer) cleanerLoop() {
	defer d.wg.Done()

	t, c := timerTestHarness(d.cleanInterval)

	select {
	case <-c:
		d.Clean()
		t.Reset(d.cleanInterval)
	case <-d.quit:
		t.Stop()
		return
	}
}

// Clean checks all session files on disk to see if they are older than
// maxAge by checking their modtime. If it finds an expired session file
// it will remove it from disk.
func (d *DiskStorer) Clean() {
	t := time.Now().UTC()

	d.mut.RLock()
	files, err := afero.ReadDir(fs, d.folderPath)
	d.mut.RUnlock()

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		// File is expired
		if file.ModTime().UTC().Add(d.maxAge).Before(t) {
			filePath := path.Join(d.folderPath, file.Name())

			d.mut.Lock()
			_, err := fs.Stat(filePath)
			// If the file has been deleted manually from the server
			// in between the time we read the directory and now, it will
			// fail here with a ErrNotExist. If so, continue gracefully.
			if os.IsNotExist(err) {
				d.mut.Unlock()
				continue
			} else if err != nil {
				panic(err)
			}

			err = fs.Remove(filePath)
			d.mut.Unlock()
			if err != nil {
				panic(err)
			}
		}
	}
}
