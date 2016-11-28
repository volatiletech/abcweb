package sessions

// DiskSession is a session storer implementation for saving sessions
// to disk.
type DiskSession struct {
	folderPath string
}

// New initializes and returns a new DiskSession.
func (d DiskSession) New(folderPath string) (DiskSession, error) {

	return DiskSession{}, nil
}
