package server

import (
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// AuthProviderStaticFile implements the AuthProvider interface that loads
// authentication keypairs into memory from a yaml file on disk. If the file is
// changed on disk, we reload the keys into memory.
type AuthProviderStaticFile struct {
	File         string
	pollInterval time.Duration
	credentials  map[string]Creds
	modTime      time.Time
	lock         sync.RWMutex
	logger       log.Logger
}

// NewAuthProviderStaticFile creates a new static file auth provider. filename is
// the path to the file on disk. pollInterval is how frequently to check the file
// for modification time updates.
func NewAuthProviderStaticFile(l log.Logger, filename string, pollInterval time.Duration) (*AuthProviderStaticFile, error) {
	a := &AuthProviderStaticFile{
		File:         filename,
		credentials:  make(map[string]Creds),
		pollInterval: pollInterval,
		logger:       l,
	}

	err := a.loadFile()
	if err != nil {
		return nil, errors.Wrap(err, "loading auth file")
	}

	return a, nil
}

// Run starts the file watcher. To stop it, close the done channel
func (a *AuthProviderStaticFile) Run(done <-chan struct{}) {
	a.logger.Log("msg", "running")
	ticker := time.NewTicker(a.pollInterval)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := a.checkReload()
			if err != nil {
				a.logger.Log("msg", "check_reload_error", "error", err)
			}
		}
	}
}

// checkReload checks the modification time of the auth file on disk. If it's different
// then we reload the auth file
func (a *AuthProviderStaticFile) checkReload() error {
	info, err := os.Stat(a.File)
	if err != nil {
		return errors.Wrap(err, "stat'ing auth file")
	}

	ts := info.ModTime()
	if ts != a.modTime {
		a.logger.Log("msg", "auth_file_reload")
		err = a.loadFile()
		if err != nil {
			return errors.Wrap(err, "reloading auth file")
		}
	}
	return nil
}

// CredsForKey given an access key, returns the set of corresponding credentials.
// If no corresponding credentials can be found, an ErrCredentialsNotFound will
// be returned.
func (a *AuthProviderStaticFile) CredsForKey(accessKey string) (*Creds, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	v, ok := a.credentials[accessKey]
	if !ok {
		return nil, ErrCredentialsNotFound
	}

	return &v, nil
}

// loadFile loads the auth file into memory
func (a *AuthProviderStaticFile) loadFile() error {
	var config struct {
		Apikeys map[string]Creds `yaml:"apikeys"`
	}

	info, err := os.Stat(a.File)
	if err != nil {
		return errors.Wrap(err, "stat'ing file")
	}

	bytes, err := ioutil.ReadFile(a.File)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return errors.Wrap(err, "unmarshaling yaml")
	}

	a.lock.Lock()
	a.credentials = config.Apikeys
	a.modTime = info.ModTime()
	a.lock.Unlock()

	return nil
}
