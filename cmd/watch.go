package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jaschaephraim/lrserver"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// assetsDirectory is used to compute the relative directory names
	// to assets to know what gulp task to run
	assetsDirectory = "assets"
	// slowdownSeconds is how long multiple "write" requests are ignored
	// for after receiving one. This is common when watching file systems since
	// when editors save files they can generate many events. (vim on one save
	// does create, write, chmod, write, chmod), the second write is likely
	// because of gofmt
	slowdownSeconds = 2
)

var (
	defaultIgnores = []string{
		".git",
		".svn",
		"vendor",
		assetsDirectory + "/vendor",
		assetsDirectory + "/js/vendor",
		assetsDirectory + "/css/vendor",
		assetsDirectory + "/img/vendor",
		assetsDirectory + "/audio/vendor",
		assetsDirectory + "/video/vendor",
		assetsDirectory + "/font/vendor",
		assetsDirectory + "/flash/vendor",
	}
)

var cmdWatch = &cobra.Command{
	Use:   "watch [flags]",
	Short: "Watch the app folder for changes",
	Long: `When files are changed, watch will detect this and run an appropriate
mechanism to re-compile them (go or assets). It also hosts a livereload server
that your app can hook into to refresh the page.`,
	Run: cmdWatchCobra,
}

func init() {
	cmdWatch.Flags().StringP("bind", "b", ":6060", "The address to proxy to the app")
	cmdWatch.Flags().StringP("root", "r", "", "The root folder to watch recursively for changes")
	cmdWatch.Flags().StringSliceP("ignore", "i", nil, "Folders to not watch, relative the app path")

	RootCmd.AddCommand(cmdWatch)
	viper.BindPFlags(cmdWatch.Flags())
}

type appWatcher struct {
	fsWatcher    *fsnotify.Watcher
	liveReloader *lrserver.Server

	watchedDirs map[string]struct{}
	slowdown    map[string]time.Time

	kill      chan struct{}
	reload    chan string
	recompile chan string

	Config struct {
		Bind   string
		Root   string
		Ignore []string
	}
}

func cmdWatchCobra(cmd *cobra.Command, args []string) {
	a := newAppWatcher(
		viper.GetString("bind"),
		viper.GetString("root"),
		viper.GetStringSlice("ignore"),
	)

	a.Watch()
}

func newAppWatcher(bindAddr, root string, ignore []string) *appWatcher {
	watcher := &appWatcher{
		watchedDirs: make(map[string]struct{}),

		kill:      make(chan struct{}),
		reload:    make(chan string),
		recompile: make(chan string),
	}

	watcher.Config.Bind = bindAddr
	watcher.Config.Root = root
	if len(root) == 0 {
		if wd, err := os.Getwd(); err != nil {
			panic(err)
		} else {
			watcher.Config.Root = wd
		}
	}
	watcher.Config.Ignore = append(defaultIgnores, ignore...)

	return watcher
}

func (a *appWatcher) Watch() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create fsnotify watcher:", err)
		return
	}
	a.fsWatcher = watcher
	defer watcher.Close()

	reloadServer := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)
	a.liveReloader = reloadServer

	wg := sync.WaitGroup{}
	wg.Add(2)

	/*go func() {
		if err = a.runApp(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to run application:", err)
		}
		wg.Done()
	}()*/
	go func() {
		if err = reloadServer.ListenAndServe(); err != nil {
			fmt.Fprintln(os.Stderr, "failed to create reloader:", err)
			panic(err)
		}
	}()
	go func() {
		if err := a.reloadWatch(); err != nil {
			fmt.Fprintln(os.Stderr, "livereload proc failed:", err)
		}
		wg.Done()
	}()
	go func() {
		if err := a.fsWatch(); err != nil {
			fmt.Fprintln(os.Stderr, "fswatch proc failed:", err)
		}
		wg.Done()
	}()

	<-sigs
	fmt.Println("Got signal, exiting")
	close(a.kill)
	wg.Wait()

	fmt.Println("\nStopped watches")
}

type statusCodeFinder interface {
	ExitStatus() int
}

func (a *appWatcher) runApp() error {
	binName := fmt.Sprintf("%s.watch", cnf.AppName)
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	var appCommand *exec.Cmd
	buildCommand := func() {
		appCommand := exec.Command(binName)
		appCommand.Stderr = os.Stderr
		appCommand.Stdout = os.Stdout
	}

	buildCommand()
	defer func() {
		if !appCommand.ProcessState.Exited() {
			appCommand.Process.Kill()
		}
	}()

	for {
		select {
		case <-a.recompile:
			output, success, err := buildApplication(binName)
			if err != nil {
				fmt.Fprintln(os.Stderr, "failed to build the application", err)
				break
			} else if !success {
				fmt.Fprintln(os.Stderr, "failed to build the application:", output)
				break
			}

			if !appCommand.ProcessState.Exited() {
				if err := appCommand.Process.Kill(); err != nil {
					fmt.Fprintln(os.Stderr, "failed to kill the application:", err)
				}
				if err := appCommand.Wait(); err != nil {
					fmt.Fprintln(os.Stderr, "failed to wait for the application:", err)
				}
			}

			if err = appCommand.Start(); err != nil {
				fmt.Fprintln(os.Stderr, "failed to start the application:", err)
			}
		case <-a.kill:
			break
		}
	}
}

// fsWatch watches files, and signals handlers when things change in those
// files.
func (a *appWatcher) fsWatch() error {
	// BUG(aarondl): thisRenameRemoves is a hack because of the insanity
	// involved in the fsnotify interface when you rename something. First
	// you get a fsnotify.Rename op specifying the old thing, then you get
	// a fsnotify.Create with the new thing, then you get another fsnotify.Rename
	// with the old thing. This simply blocks the second removal so we don't
	// get a slew of errors and do something for no reason.
	thisRenameRemoves := true

	if err := a.addRecursiveWatch(a.Config.Root); err != nil {
		return err
	}

WatchLoop:
	for {
		select {
		case e := <-a.fsWatcher.Events:
			fmt.Println("op:", e.String())
			switch e.Op {
			case fsnotify.Create:
				if err := a.watchOnCreate(e.Name); err != nil {
					fmt.Fprintln(os.Stderr, err)
					break
				}

			case fsnotify.Remove, fsnotify.Rename:
				if e.Op == fsnotify.Rename {
					doRemove := thisRenameRemoves
					thisRenameRemoves = !thisRenameRemoves
					if !doRemove {
						break
					}
				}

				if err := a.watchOnRemove(e.Name); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			case fsnotify.Write:
				if err := a.watchOnWrite(e.Name); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
		case e := <-a.fsWatcher.Errors:
			fmt.Fprintln(os.Stderr, "error during watching:", e.Error())
		case <-a.kill:
			break WatchLoop
		}
	}

	if err := a.fsWatcher.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "error during watcher close:", err)
	}

	return nil
}

func (a *appWatcher) reloadWatch() error {
WatchLoop:
	for {
		select {
		case path := <-a.reload:
			a.liveReloader.Reload(path)
		case <-a.kill:
			break WatchLoop
		}
	}

	return nil
}

func (a *appWatcher) addRecursiveWatch(path string) error {
	dirs, err := findAllDirs(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read directories", path)
	}

	fmt.Println("before ignore", dirs)
	dirs = removeIgnored(a.Config.Root, dirs, a.Config.Ignore)
	fmt.Println("after ignore", dirs)

	fmt.Println("Watching:")
	for _, d := range dirs {
		fmt.Printf(" %s\n", d)
		if err = a.fsWatcher.Add(d); err != nil {
			fmt.Fprintf(os.Stderr, "failed to add watch: %s, %v\n", a, err)
		} else {
			a.watchedDirs[d] = struct{}{}
		}
	}

	return nil
}

func (a *appWatcher) watchOnCreate(path string) error {
	fmt.Println("create:", path)
	stat, err := appFS.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat newly created file %s", path)
	}

	if !stat.IsDir() {
		return a.watchOnWrite(path)
	}

	return a.addRecursiveWatch(path)
}

func (a *appWatcher) watchOnRemove(path string) error {
	fmt.Println("remove:", path)

	toRemove := filepath.Clean(path)
	if len(toRemove) == 0 || toRemove == "." {
		return nil
	}

	for dir := range a.watchedDirs {
		dir = filepath.Clean(dir)

		if strings.HasPrefix(dir, toRemove) {
			delete(a.watchedDirs, dir)
			// BUG(aarondl): https://github.com/fsnotify/fsnotify/issues/195
			// Deadlock occurs when removing a watch from within a handler, so we launch goroutines
			// in the hopes that it'll not deadlock and finish, and we throw error handling and
			// determinism into the wind, hurray!
			go func(toRemove string) {
				fmt.Println("removing watch:", dir)
				if err := a.fsWatcher.Remove(toRemove); err != nil {
					fmt.Fprintf(os.Stderr, "failed to remove watch: %s, %v\n", dir, err)
				}
			}(dir)
		}
	}
	fmt.Println("watched dirs:", a.watchedDirs)
	return nil
}

func (a *appWatcher) watchOnWrite(path string) error {
	fmt.Println("write:", path)

	now := time.Now()
	if t, ok := a.slowdown[path]; ok {
		if int(now.Sub(t).Seconds()) < slowdownSeconds {
			return nil
		}
	}
	a.slowdown[path] = now

	stat, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat modified file %s", path)
	}

	if stat.IsDir() {
		return nil
	}

	dir := filepath.Dir(path)
	toks := strings.Split(dir, string(filepath.Separator))

	var reload bool
	switch {
	case filepath.Ext(path) == ".go":
		a.recompile <- path
	case len(toks) > 0 && toks[0] == assetsDirectory:
		reload, err = callGulp(path)
	}

	if err != nil {
		return err
	}

	if reload {
		a.reload <- path
	}

	return nil
}

func buildApplication(binName string) (string, bool, error) {
	fmt.Println("building\n")
	b := &bytes.Buffer{}
	cmd := exec.Command("go", "build", "-o", binName)
	cmd.Dir = cnf.AppPath
	cmd.Stdout = b
	cmd.Stderr = b

	err := cmd.Run()
	return b.String(), cmd.ProcessState.Success(), err
}

func callGulp(path string) (bool, error) {
	dir, _ := filepath.Split(path)

	toks := strings.Split(dir, string(filepath.Separator))
	if len(toks) < 2 {
		return false, nil
	}

	fmt.Println("gulping\n")
	cmd := exec.Command("gulp", toks[1])
	cmd.Dir = cnf.AppPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return err != nil, err
}

func findAllDirs(dir string) ([]string, error) {
	dirSet := map[string]struct{}{}
	err := afero.Walk(appFS, dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		dirSet[path] = struct{}{}
		return nil
	})

	if err != nil {
		return nil, err
	}

	var dirs []string
	for d := range dirSet {
		dirs = append(dirs, d)
	}

	return dirs, nil
}

func removeIgnored(root string, dirs []string, ignore []string) []string {
	var newDirs []string

	for _, d := range dirs {
		ignoreCheck, err := filepath.Rel(root, d)
		if err != nil {
			ignoreCheck = d
		}

		found := false
		for _, i := range ignore {
			if strings.HasPrefix(ignoreCheck, i) {
				found = true
				break
			}
		}

		if found {
			continue
		}

		newDirs = append(newDirs, d)
	}

	return newDirs
}
