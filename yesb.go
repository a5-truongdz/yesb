package main

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
const (
    dfsUnvisited = iota
    dfsVisiting
    dfsDone
)

// Modified helpers
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

// Log helpers
func fatal(format string, args ...any) {
    msg := fmt.Sprintf(format, args...)
    fmt.Printf("\033[31;1m[error]\033[0;1m: %s\n\033[0m", msg)
}

func info(format string, args ...any) {
    msg := fmt.Sprintf(format, args...)
    fmt.Printf("\033[36;1m[info]\033[0;1m: %s\n\033[0m", msg)
}

func execute(format string, args ...any) {
    msg := fmt.Sprintf(format, args...)
    fmt.Printf("\033[32;1m[execute]\033[0;1m: %s\n\033[0m", msg)
}

func skip(format string, args ...any) {
    msg := fmt.Sprintf(format, args...)
    fmt.Printf("\033[33;1m[skip]\033[0;1m: %s\n\033[0m", msg)
}

func timetaken(format string, args ...any) {
    msg := fmt.Sprintf(format, args...)
    fmt.Printf("\033[35;1m[time]\033[0;1m: %s\n\033[0m", msg)
}

// Stolen from Tsoding's nob.h  aka. "Go Rebuild Urself"
// https://github.com/tsoding/nob.h
func GoRebuildUrself() {
    exe, err := os.Executable()
    if err != nil || strings.HasPrefix(exe, "/tmp") {    // ignoring /tmp/* cuz why not
        fatal("executable not found")
        fatal("maybe you should `go build` instead of `go run`.")
        os.Exit(69)
    }

    for _, file := range []string{"build.go", "yesb.go"} {
        modified, err := isModified(file, exe)
        if err != nil {
            fatal("failed to stat `%s`: `%s`", file, err)
            fatal("make sure you name your file correctly.")
            os.Exit(69)
        }

        if modified {
            info("%s is modified", file)
            info("rebuilding it...")

            cmd := exec.Command(
                "go", "build",
                "build.go",
                "yesb.go",
            )

            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr

            if err := cmd.Run(); err != nil {
                fatal("failed to rebuild: `%s`.", err)
                fatal("make sure yesb.go exists.")
                os.Exit(69)
            }

            fmt.Println()    // stray newline for readability
            if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
                fatal("failed to restart: `%s`.", err)
                os.Exit(69)
            }
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
    uses []string
    output string
    outputFlag string
    flags []string
    alwaysRun bool
}

func NewBuildCmd() *BuildCmd {
    return &BuildCmd{}
}

func (c *BuildCmd) UseExecutable(binary string) {
    c.executable = binary
}

func (c *BuildCmd) WillUse(files ...string) {
    c.uses = append(c.uses, files...)
}

func (c *BuildCmd) WillOutput(file string) {
    c.output = file
}

func (c *BuildCmd) OutputFlag(flag string) {
    c.outputFlag = flag
}

func (c *BuildCmd) UseFlags (flags ...string) {
    c.flags = append(c.flags, flags...)
}

func (c *BuildCmd) AlwaysRun(state bool) {
    c.alwaysRun = state
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

    cmd := exec.Command(c.executable, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd
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

// Manually specify the full command
type BuildCmdManually struct {
    executable string
    args []string
}

func NewBuildCmdManually() *BuildCmdManually {
    return &BuildCmdManually{}
}

func (c *BuildCmdManually) UseExecutable(binary string) {
    c.executable = binary
}

func (c *BuildCmdManually) UseArguments(args ...string) {
    c.args = append(c.args, args...)
}

func (c *BuildCmdManually) construct() *exec.Cmd {
    cmd := exec.Command(c.executable, c.args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    return cmd
}

func (c *BuildCmdManually) run() error {
    return c.construct().Run()
}

func (c *BuildCmdManually) modified() (bool, error) {
    // well we dont know the input/output of custom commands
    // so always run it
    return true, nil
}


type BuildTarget struct {
    cmds []cmd
    dependsOn []*BuildTarget
    name string
}

func NewBuildTarget() *BuildTarget {
    return &BuildTarget{}
}

func (t *BuildTarget) UseName(name string) {
    t.name = name
}

func (t *BuildTarget) UseCommands(command ...cmd) {
    t.cmds = append(t.cmds, command...)
}

func (t *BuildTarget) DependsOn(target ...*BuildTarget) {
    t.dependsOn = append(t.dependsOn, target...)
}


type Builder struct {
    targets []*BuildTarget
    state map[*BuildTarget]int
    path []*BuildTarget
}

func NewBuilder() *Builder {
    return &Builder{}
}

func (b *Builder) UseTargets(target ...*BuildTarget) {
    b.targets = append(b.targets, target...)
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
    info("reaching target `%s`", target.name)

    for _, cmd := range target.cmds {
        cmdString := strings.Join(cmd.construct().Args, " ")

        // skip when unmodified
        modified, err := cmd.modified()
        if err != nil {
            fatal("failed to check command `%s`: `%s`", cmdString, err)
            os.Exit(69)
        }

        if !modified {
            skip("`%s` (up to date)\n", cmdString)
            continue
        }

        execute("`%s`", cmdString)

        cmdStart := time.Now()
        if err := cmd.run(); err != nil {
            fatal("failed to build target `%s`: `%s`", target.name, err)
            os.Exit(69)
        }
        cmdEnd := time.Since(cmdStart).Round(time.Microsecond)

        timetaken("command took %v\n", cmdEnd)
    }

    return nil
}

// Build() should be called only ONCE
func (b *Builder) Build() {
    b.state = make(map[*BuildTarget]int)
    b.path = nil

    buildStart := time.Now()
    for _, target := range b.targets {
        err := b.buildTarget(target)
        if err != nil {
            fatal(err.Error())
            os.Exit(69)
        }
    }
    buildEnd := time.Since(buildStart).Round(time.Millisecond)

    timetaken("build succeed, took %v", buildEnd)
}
