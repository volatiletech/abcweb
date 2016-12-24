package app

import "github.com/nullbio/abcweb/sessions"

// InitSessions initializes the sessions Overseer
func (s State) InitSessions() {
	// Configure cookie options
	opts := sessions.NewCookieOptions()
	// If not using HTTPS, disable cookie secure flag
	if len(s.Config.TLSBind) == 0 {
		opts.Secure = false
	}

	if s.Config.SessionsDevStorer {
		s.Session = sessions.NewCookieOverseer(opts, []byte("GenerateSecretKeyHere"))
	} else {
		storer, err := sessions.NewDefaultDiskStorer("GenerateUniqueTmpSubFolderHere")
		if err != nil {
			panic(err)
		}
		storer.StartCleaner()
		s.Session = sessions.NewStorageOverseer(opts, storer)
	}
}
