package processor

import (
  "os"
  "fmt"
  "time"
  "reflect"
  "path/filepath"
  "github.com/fsnotify/fsnotify"
)

type ChangeCode int
type TimeoutErr struct{}
func (t *TimeoutErr) Error() string {
  return "Experienced time out"
}
var NextChangeTimeout time.Duration = 0

type fsDetector struct {
  watchers []*fsnotify.Watcher
  cases []reflect.SelectCase
  keymap map[int]string
  closed bool
}

func newFsDetector() *fsDetector {
  return &fsDetector{
    watchers: make([]*fsnotify.Watcher, 0),
    cases: make([]reflect.SelectCase, 0),
    keymap: make(map[int]string),
    closed: false,
  }
}

func (f *fsDetector) Watch(root string) error {
  if f.closed {
    return fmt.Errorf("fsDetector is closed")
  }

  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    return fmt.Errorf("Couldn't retrieve new watcher in fsDetector.Watch(): %v", err)
  }

  err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
  	if fi.Mode().IsDir() {
  		return watcher.Add(path)
  	}
  	return nil
  })

  if err != nil {
    return fmt.Errorf("Couldn't walk %s in watchDirectory(): %v", root, err)
  }

  f.watchers = append(f.watchers, watcher)
  newCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(watcher.Events)}
  f.cases = append(f.cases, newCase)
  f.keymap[len(f.watchers) - 1] = root
  return nil
}

func (f *fsDetector) Unwatch(root string) error {
  if f.closed {
    return fmt.Errorf("fsDetector is closed")
  }

  watcherIdx := -1
  for key, val := range f.keymap {
    if val == root {
      watcherIdx = key
      break
    }
  }

  if watcherIdx == -1 {
    return fmt.Errorf("No watch on %s in fsDetector.Unwatch()", root)
  }
  delete(f.keymap, watcherIdx)

  watcher := f.watchers[watcherIdx]
  if watcherIdx == len(f.watchers) - 1 {
    f.watchers = f.watchers[:watcherIdx]
    f.cases = f.cases[:watcherIdx]
  } else {
    f.watchers = append(f.watchers[:watcherIdx], f.watchers[watcherIdx+1:]...)
    f.cases = append(f.cases[:watcherIdx], f.cases[watcherIdx+1:]...)
  }

  watcher.Close()
  return nil
}

func (f *fsDetector) NextChange() (string, error) {
  if f.closed {
    return "", fmt.Errorf("fsDetector is closed")
  }

  cases := f.cases
  if NextChangeTimeout > 0 {
    timeoutCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(time.After(NextChangeTimeout))}
    cases = append(cases, timeoutCase)
  }
  chosen, _, ok := reflect.Select(cases)
  if !ok {
    return "", fmt.Errorf("Failed to select value in fsDetector.NextChange()")
  } else if chosen == len(f.cases) {
    return "", &TimeoutErr{}
  }
  return f.keymap[chosen], nil
}

func (f *fsDetector) Close() {
  for _, watcher := range f.watchers {
    watcher.Close()
  }
  f.closed = true
}
