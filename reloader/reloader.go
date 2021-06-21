package reloader

import (
	"context"
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var ErrUnknown = errors.New("unknown error")

type Reloader struct {
	dir  string
	load func(files map[string][]byte)
}

func New(dir string, load func(files map[string][]byte)) (*Reloader, error) {
	r := &Reloader{
		dir:  dir,
		load: load,
	}
	err := os.MkdirAll(dir, os.ModeDir|0755)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Reloader) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	err = watcher.Add(r.dir)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return errors.New("unknown error")
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				r.LoadOnce()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return ErrUnknown
			}
			log.Print("error:", err)
		}
	}
}

func (r *Reloader) LoadOnce() {
	r.load(r.files())
}

func (r *Reloader) files() map[string][]byte {
	m := map[string][]byte{}
	_ = filepath.WalkDir(r.dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}
		m[path] = data
		return nil
	})
	return m
}
