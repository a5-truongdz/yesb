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

// enums for DFS
type dfsState int
const (
    dfsUnvisited dfsState = iota
    dfsVisiting
    dfsDone
)

// modified helper
// compare if file is newer than ref, which assumes modified
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

// command helper
func constructCmd(executable string, args ...string) *exec.Cmd {
    cmd := exec.Command(executable, args...)

    cmd.Stdout = os.Stdout
    cmd.Stdin  = os.Stdin
    cmd.Stderr = os.Stderr

    return cmd
}

// log helpers
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
    beautify(34, "skip", format, args...)
}

func logTime(format string, args ...any) {
    beautify(35, "time", format, args...)
}

func logWarn(format string, args ...any) {
    beautify(33, "warn", format, args...)
}

// stolen from Tsoding's nob.h  aka. "Go Rebuild Urself"
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
            "-o", exe,
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
    construct() *exec.Cmd
    run() error
    modified() (bool, error)
}

// {executable} {flags} {use} {outputFlag} {output}
type BuildCmd struct {
    executable string
    uses       []string
    output     string
    outputFlag string
    flags      []string
    alwaysRun  bool
}

func NewBuildCmd() *BuildCmd {
    return &BuildCmd{}
}

func (c *BuildCmd) UseExecutable(binary string) *BuildCmd {
    c.executable = binary
    return c
}

func (c *BuildCmd) WillUse(files ...string) *BuildCmd {
    c.uses = append(c.uses, files...)
    return c
}

func (c *BuildCmd) WillOutput(file string) *BuildCmd {
    c.output = file
    return c
}

func (c *BuildCmd) OutputFlag(flag string) *BuildCmd {
    c.outputFlag = flag
    return c
}

func (c *BuildCmd) UseFlags (flags ...string) *BuildCmd {
    c.flags = append(c.flags, flags...)
    return c
}

func (c *BuildCmd) AlwaysRun(state bool) *BuildCmd {
    c.alwaysRun = state
    return c
}

func (c *BuildCmd) construct() *exec.Cmd {
    var args []string

    args = append(args, c.flags...)
    args = append(args, c.uses...)

    // outputFlags can be empty
    if c.outputFlag != "" {
        args = append(args, c.outputFlag)
    }

    // again, output can be empty
    if c.output != "" {
        args = append(args, c.output)
    }

    return constructCmd(c.executable, args...)
}

func (c *BuildCmd) run() error {
    return c.construct().Run()
}

func (c *BuildCmd) modified() (bool, error) {
    if c.alwaysRun {
        return true, nil
    }

    // output not specified
    if c.output == "" {
        logWarn("output not specified")
        logWarn("running anyways...")
        return true, nil
    }

    // output isnt generated yet
    if _, err := os.Stat(c.output); errors.Is(err, os.ErrNotExist) {
        return true, nil
    }

    for _, file := range c.uses {
        modified, err := isModified(file, c.output)
        if err != nil {
            return false, err
        }

        if modified {
            return true, nil
        }
    }

    return false, nil
}

// manually specify the full command
type BuildCmdManually struct {
    executable string
    args       []string
    track      bool
    files      []string
    reference  string
}

func NewBuildCmdManually() *BuildCmdManually {
    return &BuildCmdManually{}
}

func (c *BuildCmdManually) UseExecutable(binary string) *BuildCmdManually {
    c.executable = binary
    return c
}

func (c *BuildCmdManually) UseArguments(args ...string) *BuildCmdManually {
    c.args = append(c.args, args...)
    return c
}

func (c *BuildCmdManually) Track(flag bool) *BuildCmdManually {
    c.track = flag
    return c
}

func (c *BuildCmdManually) TrackFiles(files ...string) *BuildCmdManually {
    c.files = append(c.files, files...)
    return c
}

func (c *BuildCmdManually) TrackReference(file string) *BuildCmdManually {
    c.reference = file
    return c
}

func (c *BuildCmdManually) construct() *exec.Cmd {
    return constructCmd(c.executable, c.args...)
}

func (c *BuildCmdManually) run() error {
    return c.construct().Run()
}

func (c *BuildCmdManually) modified() (bool, error) {
    if !c.track {
        if len(c.files) != 0 {
            logWarn("tracking not set but files are specified")
        } else if c.reference != "" {
            logWarn("tracking not set but reference is specified")
        }

        return true, nil
    }

    // track set but files|reference is empty
    if len(c.files) == 0 {
        logWarn("no files specified for tracking")
        logWarn("running anyways...")
        return true, nil
    }

    if c.reference == "" {
        logWarn("no reference specified for tracking")
        logWarn("running anyways...")
        return true, nil
    }

    for _, file := range c.files {
        modified, err := isModified(file, c.reference)
        if err != nil {
            return false, err
        }

        if modified {
            return true, nil
        }
    }

    return false, nil
}


type BuildTarget struct {
    cmds      []cmd
    dependsOn []*BuildTarget
    name      string
}

func NewBuildTarget() *BuildTarget {
    return &BuildTarget{}
}

// since we dont know what the variable name is
// we have to do manually provide the name
func (t *BuildTarget) UseName(name string) *BuildTarget {
    t.name = name
    return t
}

func (t *BuildTarget) UseCommands(commands ...cmd) *BuildTarget {
    t.cmds = append(t.cmds, commands...)
    return t
}

func (t *BuildTarget) DependsOn(target ...*BuildTarget) *BuildTarget {
    t.dependsOn = append(t.dependsOn, target...)
    return t
}


type Builder struct {
    targets []*BuildTarget
    state   map[*BuildTarget]dfsState
    path    []*BuildTarget
}

func NewBuilder() *Builder {
    return &Builder{}
}

func (b *Builder) UseTargets(target ...*BuildTarget) *Builder {
    b.targets = append(b.targets, target...)
    return b
}

func (b *Builder) buildTarget(target *BuildTarget) error {
    switch b.state[target] {
    case dfsVisiting:
        start := 0

        for i, t := range b.path {
            if t == target {
                start = i
                break
            }
        }

        var cycle []string
        for _, t := range b.path[start:] {
            cycle = append(cycle, fmt.Sprintf("`%s`", t.name))
        }
        cycle = append(cycle, fmt.Sprintf("`%s`", target.name))

        return fmt.Errorf(
            "dependency cycle detected: %s",
            strings.Join(cycle, " -> "),
        )

    case dfsDone:
        return nil
    }

    b.state[target] = dfsVisiting
    b.path = append(b.path, target)

    // build deps first
    for _, dep := range target.dependsOn {
        err := b.buildTarget(dep)
        if err != nil {
            return err
        }
    }

    b.state[target] = dfsDone
    b.path = b.path[:len(b.path) - 1]

    // then build target
    logInfo("reaching target `%s`", target.name)

    for _, cmd := range target.cmds {
        cmdString := strings.Join(cmd.construct().Args, " ")

        // skip when unmodified
        modified, err := cmd.modified()
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
        if err := cmd.run(); err != nil {
            logFatal("failed to build target `%s`: `%s`", target.name, err)
            os.Exit(69)
        }
        cmdEnd := time.Since(cmdStart).Round(time.Microsecond)

        logTime("command took %v\n", cmdEnd)
    }

    return nil
}

// Build() should be called only ONCE
func (b *Builder) Build() {
    b.state = make(map[*BuildTarget]dfsState)
    b.path = nil

    buildStart := time.Now()
    for _, target := range b.targets {
        err := b.buildTarget(target)
        if err != nil {
            logFatal("%s", err)
            os.Exit(69)
        }
    }
    buildEnd := time.Since(buildStart).Round(time.Millisecond)

    logTime("build succeed, took %v", buildEnd)
}
