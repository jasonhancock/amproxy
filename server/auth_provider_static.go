package server

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jasonhancock/go-logger"
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
	logger       *logger.L
}

// NewAuthProviderStaticFile creates a new static file auth provider. filename is
// the path to the file on disk. pollInterval is how frequently to check the file
// for modification time updates.
func NewAuthProviderStaticFile(ctx context.Context, l *logger.L, filename string, pollInterval time.Duration) (*AuthProviderStaticFile, error) {
	a := &AuthProviderStaticFile{
		File:         filename,
		credentials:  make(map[string]Creds),
		pollInterval: pollInterval,
		logger:       l,
	}

	if err := a.loadFile(); err != nil {
		return nil, fmt.Errorf("loading auth file: %w", err)
	}

	go a.run(ctx)

	return a, nil
}

// run starts the file watcher. To stop it, cancel the context.
func (a *AuthProviderStaticFile) run(ctx context.Context) {
	a.logger.Info("running")
	ticker := time.NewTicker(a.pollInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := a.checkReload()
			if err != nil {
				a.logger.LogError("check_reload_error", err)
			}
		}
	}
}

// checkReload checks the modification time of the auth file on disk. If it's different
// then we reload the auth file
func (a *AuthProviderStaticFile) checkReload() error {
	info, err := os.Stat(a.File)
	if err != nil {
		return fmt.Errorf("stat'ing auth file: %w", err)
	}

	ts := info.ModTime()
	if ts != a.modTime {
		a.logger.Info("initiating auth file reload")
		if err := a.loadFile(); err != nil {
			return fmt.Errorf("reloading auth file: %w", err)
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
		Apikeys map[string]fileCreds `yaml:"apikeys"`
	}

	fh, err := os.Open(a.File)
	if err != nil {
		return fmt.Errorf("opening %q: %w", a.File, err)
	}
	defer fh.Close()

	info, err := fh.Stat()
	if err != nil {
		return fmt.Errorf("stat'ing auth file: %w", err)
	}

	if err = yaml.NewDecoder(fh).Decode(&config); err != nil {
		return fmt.Errorf("unmarshaling yaml: %w", err)
	}

	creds := make(map[string]Creds, len(config.Apikeys))
	for k, v := range config.Apikeys {
		creds[k] = v.To()
	}

	a.lock.Lock()
	a.credentials = creds
	a.modTime = info.ModTime()
	a.lock.Unlock()

	return nil
}

type fileCreds struct {
	AccessKey string   `yaml:"access_key"`
	SecretKey string   `yaml:"secret_key"`
	Metrics   []string `yaml:"metrics"`
}

func (c fileCreds) To() Creds {
	cr := Creds{
		SecretKey: c.SecretKey,
		Metrics:   make(map[string]struct{}, len(c.Metrics)),
	}

	for _, v := range c.Metrics {
		cr.Metrics[v] = struct{}{}
	}

	return cr
}
