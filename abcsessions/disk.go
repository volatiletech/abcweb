package abcsessions

import (
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/djherbis/times"
	"github.com/pkg/errors"
)

// DiskStorer is a session storer implementation for saving sessions
// to disk.
type DiskStorer struct {
	// Path to the session files folder
	folderPath string
	// How long sessions take to expire on disk
	// Note that this is seperate to the cookie maxAge
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

// NewDefaultDiskStorer returns a DiskStorer object with default values.
// The default values are:
// FolderPath: system tmp dir + random folder
// maxAge: 2 days (clear session stored on server after 2 days)
// cleanInterval: 1 hour (delete sessions older than maxAge every 1 hour)
func NewDefaultDiskStorer(tmpSubFolder string) (*DiskStorer, error) {
	folderPath := path.Join(os.TempDir(), tmpSubFolder)
	return NewDiskStorer(folderPath, time.Hour*24*2, time.Hour)
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
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		err := os.Mkdir(folderPath, os.FileMode(int(0755)))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to make directory: %s", folderPath)
		}
	}

	return d, nil
}

// All keys in the disk store
func (d *DiskStorer) All() ([]string, error) {
	files, err := ioutil.ReadDir(d.folderPath)
	if err != nil {
		return []string{}, errors.Wrapf(err, "unable to read directory: %s", d.folderPath)
	}

	sessions := make([]string, len(files))

	for i := 0; i < len(files); i++ {
		sessions[i] = files[i].Name()
	}

	return sessions, nil
}

// Get returns the value string saved in the session pointed to by the
// session id key.
func (d *DiskStorer) Get(key string) (value string, err error) {
	if !validKey(key) {
		return "", errNoSession{}
	}

	filePath := path.Join(d.folderPath, key)

	d.mut.RLock()
	defer d.mut.RUnlock()

	_, err = os.Stat(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to stat session file: %s", filePath)
	}

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read file: %s", filePath)
	}

	return string(contents), nil
}

// Set saves the value string to the session pointed to by the session id key.
func (d *DiskStorer) Set(key, value string) error {
	if !validKey(key) {
		return errNoSession{}
	}

	filePath := path.Join(d.folderPath, key)

	d.mut.Lock()
	defer d.mut.Unlock()

	return ioutil.WriteFile(filePath, []byte(value), 0600)
}

// Del the session pointed to by the session id key and remove it.
func (d *DiskStorer) Del(key string) error {
	if !validKey(key) {
		return errNoSession{}
	}

	filePath := path.Join(d.folderPath, key)

	d.mut.Lock()
	defer d.mut.Unlock()

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "unable to stat session file: %s", filePath)
	}

	return os.Remove(filePath)
}

// StopCleaner stops the cleaner go routine
func (d *DiskStorer) StopCleaner() {
	close(d.quit)
	d.wg.Wait()
}

// ResetExpiry resets the expiry of the key
func (d *DiskStorer) ResetExpiry(key string) error {
	if !validKey(key) {
		return errNoSession{}
	}

	filePath := path.Join(d.folderPath, key)
	nowTime := time.Now().UTC()

	return os.Chtimes(filePath, nowTime, nowTime)
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
// maxAge by checking their access time. If it finds an expired session file
// it will remove it from disk.
func (d *DiskStorer) Clean() {
	t := time.Now().UTC()

	files, err := ioutil.ReadDir(d.folderPath)

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		tspec := times.Get(file)

		// File is expired
		if tspec.AccessTime().UTC().UTC().Add(d.maxAge).Before(t) {
			filePath := path.Join(d.folderPath, file.Name())

			d.mut.Lock()
			_, err := os.Stat(filePath)
			// If the file has been deleted manually from the server
			// in between the time we read the directory and now, it will
			// fail here with a ErrNotExist. If so, continue gracefully.
			// It would be innefficient to hold a lock for the duration of
			// the loop, so we only lock when we find an expired file.
			if os.IsNotExist(err) {
				d.mut.Unlock()
				continue
			} else if err != nil {
				panic(err)
			}

			err = os.Remove(filePath)
			d.mut.Unlock()
			if err != nil {
				panic(err)
			}
		}
	}
}
