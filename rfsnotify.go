package rfsnotify

import (
	"github.com/fsnotify/fsnotify"
)

type Event = fsnotify.Event

type Op = fsnotify.Op

const (
	Create fsnotify.Op = fsnotify.Create
	Write  fsnotify.Op = fsnotify.Write
	Remove fsnotify.Op = fsnotify.Remove
	Rename fsnotify.Op = fsnotify.Rename
	Chmod  fsnotify.Op = fsnotify.Chmod
)

// Common errors that can be reported by a watcher
var (
	ErrNonExistentWatch = fsnotify.ErrNonExistentWatch
	ErrEventOverflow    = fsnotify.ErrEventOverflow
)
