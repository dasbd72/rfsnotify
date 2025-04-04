# rfsnotify

A wrapper of golang fsnotify that automatically recursively watches directory

## Usage

```go
package main

import (
 "log"

 "github.com/dasbd72/rfsnotify"
)

func main() {
    watcher, err := rfsnotify.NewWatcher()
 if err != nil {
  log.Fatal(err)
 }
 err = watcher.Add(root_dir)
 if err != nil {
  log.Fatal(err)
 }

    for {
  select {
  case ev := <-watcher.Events:
   log.Println("event:", ev)
  case err := <-watcher.Errors:
   log.Println("error:", err)
  }
  log.Println(watcher.WatchList())
 }
}
```
