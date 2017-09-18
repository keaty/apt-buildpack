package apt

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Command interface {
	// Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(string, string, ...string) (string, error)
	// Run(*exec.Cmd) error
}

type Apt struct {
	command    Command
	options    []string
	aptFile    string
	cacheDir   string
	stateDir   string
	installDir string
}

func New(command Command, aptFile, cacheDir, installDir string) *Apt {
	s := &Apt{
		command:    command,
		aptFile:    aptFile,
		cacheDir:   filepath.Join(cacheDir, "apt", "cache"),
		stateDir:   filepath.Join(cacheDir, "apt", "state"),
		installDir: installDir,
	}
	s.options = []string{
		"-o", "debug::nolocking=true",
		"-o", "dir::cache=" + s.cacheDir,
		"-o", "dir::state=" + s.stateDir,
	}
	return s
}

func (a *Apt) Update() (string, error) {
	args := append(a.options, "update")
	return a.command.Output("/", "apt-get", args...)
}

func (a *Apt) Download() (string, error) {
	aptArgs := append(a.options, "-y", "--force-yes", "-d", "install", "--reinstall")

	text, err := ioutil.ReadFile(a.aptFile)
	if err != nil {
		return "", err
	}
	for _, pkg := range strings.Split(string(text), "\n") {
		if strings.HasSuffix(pkg, ".deb") {
			packageFile := filepath.Join(a.cacheDir, "archives", filepath.Base(pkg))
			args := []string{"-s", "-L", "-z", packageFile, "-o", packageFile, pkg}
			if output, err := a.command.Output("/", "curl", args...); err != nil {
				return output, err
			}

		} else if pkg != "" {
			args := append(aptArgs, pkg)
			if output, err := a.command.Output("/", "apt-get", args...); err != nil {
				return output, err
			}
		}
	}

	return "", nil
}

func (a *Apt) Install() (string, error) {
	files, err := filepath.Glob(filepath.Join(a.cacheDir, "archives", "*.deb"))
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if output, err := a.command.Output("/", "dpkg", "-x", file, filepath.Join(a.installDir, "apt")); err != nil {
			return output, err
		}
	}
	return "", nil
}
