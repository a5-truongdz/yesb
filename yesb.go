package yesb

import (
    "os"
    "os/exec"
    "fmt"
    "syscall"
    "strings"
    "time"
    "errors"
)

// Enums for DFS
type dfsState int
const (
    dfsUnvisited dfsState = iota
    dfsVisiting
    dfsDone
)

// Modified helper
// Compare if file is newer than ref, which assumes modified
func isModified(file string, ref string) (bool, error) {
    fileInfo, err := os.Stat(file)
    if err != nil {
        return false, err
    }

    refInfo, err := os.Stat(ref)
    if err != nil {
        return false, err
    }

    return fileInfo.ModTime().After(refInfo.ModTime()), nil
}

// Command helper
func constructCmd(executable string, args ...string) *exec.Cmd {
    cmd := exec.Command(executable, args...)

    cmd.Stdout = os.Stdout
    cmd.Stdin  = os.Stdin
    cmd.Stderr = os.Stderr

    return cmd
}

// Log helpers
func beautify(color int, tag, format string, args ...any) {
    msg := fmt.Sprintf(format, args...)
    fmt.Printf("\033[%d;1m[%s]\033[0;1m: %s\n\033[0m", color, tag, msg)
}

func logFatal(format string, args ...any) {
    beautify(31, "error", format, args...)
}

func logInfo(format string, args ...any) {
    beautify(36, "info", format, args...)
}

func logExecute(format string, args ...any) {
    beautify(32, "execute", format, args...)
}

func logSkip(format string, args ...any) {
    beautify(33, "skip", format, args...)
}

func logTime(format string, args ...any) {
    beautify(35, "time", format, args...)
}

// Stolen from Tsoding's nob.h  aka. "Go Rebuild Urself"
// https://github.com/tsoding/nob.h
func GoRebuildUrself() {
    exe, err := os.Executable()
    if err != nil || strings.HasPrefix(exe, "/tmp") {    // ignoring /tmp/* cuz why not
        logFatal("executable not found")
        logFatal("maybe you should `go build` instead of `go run`.")
        os.Exit(69)
    }

    file := "build.go"
    modified, err := isModified(file, exe)
    if err != nil {
        logFatal("failed to stat `%s`: `%s`", file, err)
        logFatal("make sure you name your file correctly.")
        os.Exit(69)
    }

    if modified {
        logInfo("%s is modified", file)
        logInfo("rebuilding it...")

        cmd := constructCmd(
            "go", "build",
            "-tags", "build",
            "build.go",
        )

        if err := cmd.Run(); err != nil {
            logFatal("failed to rebuild: `%s`.", err)
            logFatal("make sure yesb.go exists.")
            os.Exit(69)
        }

        fmt.Println()    // stray newline for readability
        if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
            logFatal("failed to restart: `%s`.", err)
            os.Exit(69)
        }
    }
}


type cmd interface {
    _construct() *exec.Cmd
    _run() error
    _modified() (bool, error)
}

// {executable} {flags} {use} {outputFlag} {output}
type BuildCmd struct {
    _executable string
    _uses []string
    _output string
    _outputFlag string
    _flags []string
    _alwaysRun bool
}

func NewBuildCmd() *BuildCmd {
    return &BuildCmd{}
}

func (c *BuildCmd) UseExecutable(binary string) {
    c._executable = binary
}

func (c *BuildCmd) WillUse(files ...string) {
    c._uses = append(c._uses, files...)
}

func (c *BuildCmd) WillOutput(file string) {
    c._output = file
}

func (c *BuildCmd) OutputFlag(flag string) {
    c._outputFlag = flag
}

func (c *BuildCmd) UseFlags (flags ...string) {
    c._flags = append(c._flags, flags...)
}

func (c *BuildCmd) AlwaysRun(state bool) {
    c._alwaysRun = state
}

func (c *BuildCmd) _construct() *exec.Cmd {
    var args []string

    args = append(args, c._flags...)
    args = append(args, c._uses...)

    // outputFlags can be empty
    if c._outputFlag != "" {
        args = append(args, c._outputFlag)
    }

    // again, output can be empty
    if c._output != "" {
        args = append(args, c._output)
    }

    return constructCmd(c._executable, args...)
}

func (c *BuildCmd) _run() error {
    return c._construct().Run()
}

func (c *BuildCmd) _modified() (bool, error) {
    if c._alwaysRun {
        return true, nil
    }

    // output not specified
    if c._output == "" {
        return true, nil
    }

    // output isnt generated yet
    if _, err := os.Stat(c._output); errors.Is(err, os.ErrNotExist) {
        return true, nil
    }

    for _, file := range c._uses {
        modified, err := isModified(file, c._output)
        if err != nil {
            return false, err
        }

        if modified {
            return true, nil
        }
    }

    return false, nil
}

// Manually specify the full command
type BuildCmdManually struct {
    _executable string
    _args []string
}

func NewBuildCmdManually() *BuildCmdManually {
    return &BuildCmdManually{}
}

func (c *BuildCmdManually) UseExecutable(binary string) {
    c._executable = binary
}

func (c *BuildCmdManually) UseArguments(args ...string) {
    c._args = append(c._args, args...)
}

func (c *BuildCmdManually) _construct() *exec.Cmd {
    return constructCmd(c._executable, c._args...)
}

func (c *BuildCmdManually) _run() error {
    return c._construct().Run()
}

func (c *BuildCmdManually) _modified() (bool, error) {
    // well we dont know the input/output of custom commands
    // so always run it
    return true, nil
}


type BuildTarget struct {
    _cmds []cmd
    _dependsOn []*BuildTarget
    _name string
}

func NewBuildTarget() *BuildTarget {
    return &BuildTarget{}
}

// since we dont know what the variable name are
// we have to do manually provide the _name
func (t *BuildTarget) UseName(name string) {
    t._name = name
}

func (t *BuildTarget) UseCommands(commands ...cmd) {
    t._cmds = append(t._cmds, commands...)
}

func (t *BuildTarget) DependsOn(target ...*BuildTarget) {
    t._dependsOn = append(t._dependsOn, target...)
}


type Builder struct {
    _targets []*BuildTarget
    _state map[*BuildTarget]dfsState
    _path []*BuildTarget
}

func NewBuilder() *Builder {
    return &Builder{}
}

func (b *Builder) UseTargets(target ...*BuildTarget) {
    b._targets = append(b._targets, target...)
}

func (b *Builder) buildTarget(target *BuildTarget) error {
    switch b._state[target] {
    case dfsVisiting:
        start := 0

        for i, t := range b._path {
            if t == target {
                start = i
                break
            }
        }

        var cycle []string
        for _, t := range b._path[start:] {
            cycle = append(cycle, fmt.Sprintf("`%s`", t._name))
        }
        cycle = append(cycle, fmt.Sprintf("`%s`", target._name))

        return fmt.Errorf(
            "dependency cycle detected: %s",
            strings.Join(cycle, " -> "),
        )

    case dfsDone:
        return nil
    }

    b._state[target] = dfsVisiting
    b._path = append(b._path, target)

    // build deps first
    for _, dep := range target._dependsOn {
        err := b.buildTarget(dep)
        if err != nil {
            return err
        }
    }

    b._state[target] = dfsDone
    b._path = b._path[:len(b._path) - 1]

    // then build target
    logInfo("reaching target `%s`", target._name)

    for _, cmd := range target._cmds {
        cmdString := strings.Join(cmd._construct().Args, " ")

        // skip when unmodified
        modified, err := cmd._modified()
        if err != nil {
            logFatal("failed to check command `%s`: `%s`", cmdString, err)
            os.Exit(69)
        }

        if !modified {
            logSkip("`%s` (up to date)\n", cmdString)
            continue
        }

        logExecute("`%s`", cmdString)

        cmdStart := time.Now()
        if err := cmd._run(); err != nil {
            logFatal("failed to build target `%s`: `%s`", target._name, err)
            os.Exit(69)
        }
        cmdEnd := time.Since(cmdStart).Round(time.Microsecond)

        logTime("command took %v\n", cmdEnd)
    }

    return nil
}

// Build() should be called only ONCE
func (b *Builder) Build() {
    b._state = make(map[*BuildTarget]dfsState)
    b._path = nil

    buildStart := time.Now()
    for _, target := range b._targets {
        err := b.buildTarget(target)
        if err != nil {
            logFatal("%s", err)
            os.Exit(69)
        }
    }
    buildEnd := time.Since(buildStart).Round(time.Millisecond)

    logTime("build succeed, took %v", buildEnd)
}
