package rfsnotify

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Events chan fsnotify.Event

	// Errors sends any errors.
	Errors chan error

	w_fsnotify *fsnotify.Watcher // Underlying fsnotify watcher
	done       chan struct{}     // Channel for sending a "quit message" to the reader goroutine
	doneResp   chan struct{}     // Channel to respond to Close
}

func NewWatcher() (*Watcher, error) {
	w_fsnotify, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Events: make(chan fsnotify.Event),
		Errors: make(chan error),

		w_fsnotify: w_fsnotify,
		done:       make(chan struct{}),
		doneResp:   make(chan struct{}),
	}

	go w.readEvents()
	return w, nil
}

func (w *Watcher) isClosed() bool {
	select {
	case <-w.done:
		return true
	default:
		return false
	}
}

// Close removes all watches and closes the events channel.
func (w *Watcher) Close() error {
	if w.isClosed() {
		return nil
	}

	// Send 'close' signal to goroutine, and set the Watcher to closed.
	close(w.done)

	// Wait for the reader to finish so we can close the events channel
	<-w.doneResp

	return w.w_fsnotify.Close()
}

func (w *Watcher) Add(name string) error {
	if w.isClosed() {
		return errors.New("already closed")
	}
	if err := w.recursive(name, true); err != nil {
		return err
	}
	return nil
}

func (w *Watcher) Remove(name string) error {
	if w.isClosed() {
		return errors.New("already closed")
	}
	if err := w.recursive(name, false); err != nil {
		return err
	}
	return nil
}

func (w *Watcher) WatchList() []string {
	return w.w_fsnotify.WatchList()
}

func (w *Watcher) recursive(path string, isAdd bool) error {
	err := filepath.Walk(path, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if isAdd {
				if err = w.w_fsnotify.Add(walkPath); err != nil {
					return err
				}
			} else {
				if err = w.w_fsnotify.Remove(walkPath); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func (w *Watcher) eventRecursive(path string, isAdd bool, eventQueue *eventQueue) error {
	err := filepath.Walk(path, func(walkPath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if walkPath == path {
			return nil
		}
		if fi.IsDir() {
			if isAdd {
				if err = w.w_fsnotify.Add(walkPath); err != nil {
					return err
				}
			} else {
				if err = w.w_fsnotify.Remove(walkPath); err != nil {
					return err
				}
			}
		}
		// Assume the path is created
		eventQueue.push(fsnotify.Event{
			Name: walkPath,
			Op:   fsnotify.Create,
		})
		return nil
	})
	return err
}

func (w *Watcher) readEvents() {
	defer func() {
		close(w.doneResp)
		close(w.Errors)
		close(w.Events)
	}()

	for {
		// See if we have been closed.
		if w.isClosed() {
			return
		}

		select {
		case event := <-w.w_fsnotify.Events:
			w.Events <- event

			// Recursively add or remove directories
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Chmod) {
				eventQueue := newEventQueue()
				if err := w.eventRecursive(event.Name, true, eventQueue); err != nil {
					w.Errors <- err
				}
				for eventQueue.size() > 0 {
					w.Events <- eventQueue.pop()
				}
			}
		case err := <-w.w_fsnotify.Errors:
			w.Errors <- err
		case <-w.done:
			w.doneResp <- struct{}{}
			return
		}
	}
}
