# yesb Syntax
> **NOTE**: `yesb` is experimental. APIs may change.

## I. The Basic
A `yesb` build script is a normal Go program.

```go
//go:build build
package main

import "github.com/a5-truongdz/yesb"

func main() {
    yesb.GoRebuildUrself()
    yesb.NewBuilder().
        UseTargets(
            yesb.NewBuildTarget().
                UseName("hello").
                UseCommands(
                    yesb.NewBuildCmdManually().
                        UseExecutable("echo").
                        UseArguments("Hello", "World!"),
                ),
        ).
        Build()
}

```

> **NOTE**: Your build script **MUST** be named `build.go`.
>
> **NOTE**: `build.go` **MUST** be declared `package main`.
>
> **NOTE**: The `//go:build build` tag is **REQUIRED**. It's used to avoid conflicting ith another Go project, since `build.go` is declared `package main`.
>
> **NOTE**: Your project **MUSTN'T** use the same tag.
>
> **NOTE**: You should also build your project with `go build` instead of `go build .`. It can build both `build.go` and your projects (which is not the behaviour you probably wanted).

To use `yesb`:
- Add `yesb` to your project: `go get github.com/a5-truongdz/yesb`.
- Create a `build.go` and define build targets and dependencies. (See syntax below)
> Don't forget to use `yesb.GoRebuildUrself()` to make it auto-rebuild!
>
> Don't forget to add the build tag `//go:build build`, please!
- Bootstrap it with `go build -tags build -o build build.go`.

After that you can just call `./build` to make it auto-rebuild.

You can use another output file. It's fine (but I don't recommend).

## II. Auto-rebuild
Auto-rebuild is achieved by `yesb.GoRebuildUrself()`. It's intended to be called once at the beginning of `build.go`.

When called, it compares the modification time of `build.go` with the executable. If `build.go` is newer, it rebuilds itself and restarts it.

It's also not intended to be used with `go run build.go`, it makes no sense to try and rebuild an executable from `/tmp`. `yesb` will reject it.

## III. Commands
Commands are the actual things performed by `yesb`.

There are currently 2 types of commands:
- `BuildCmd`: A structured command with inputs and output.
- `BuildCmdManually`: A completely manual command.

### 1. `BuildCmd`
`BuildCmd` is the normal way to describe a build command.

It is structured as `<executable> <flags> <inputs> <output flag> <output>`.

Example:
```go
cmd := yesb.NewBuildCmd().
    UseExecutable("g++").
    UseFlags("-std=c++20", "-Wall").
    WillUse("main.cpp").
    OutputFlag("-o").
    WillOutput("program")
```

Represents `g++ -std=c++20 -Wall main.cpp -o program`.

#### a, `.UseExecutable(binary)`
Specifies the executable to run.

```go
cmd.UseExecutable("g++")
```

#### b, `.UseFlags(flag...)`
Add flags to the command.

Multiple flags are allowed.

```go
cmd.UseFlags(
    "-std=c++20",
    "-Wall",
    "-Wextra",
    ...
)
```

#### c, `.WillUse(file...)`
Specifies files used by the command.

Multiple files are allowed.

```go
cmd.WillUse(
    "hello.cpp",
    "hello.hpp",
    "main.cpp",
    ...
)
```

#### d, `.OutputFlag(flag)`
Specifies the flag placed before the output.

It can be changed depending on what the executable expects.

```go
cmd.OutputFlag("-o")    // g++
cmd.OutputFlag("-d")    // javac
```

#### e, `.WillOutput(file)`
Specifies the output produced by the command.

The output is used together with `.WillUse()` for incremental builds.

```go
cmd.WillOutput("hello")
```

#### f, `.AlwaysRun(state)`
Forces the command to run everytime, regardless of modification time.

It can be enabled or disabled.

```go
cmd.AlwaysRun(true)
cmd.AlwaysRun(false)
```

### 2. `BuildCmdManually`
Sometimes a command doesn't fit the `BuildCmd` abstraction.

For example: `go vet .`.

There is no meaningful input/output relationship to track here.

That's what `BuildCmdManually` is for, it's always executed every time.

Example:
```go
cmd := yesb.NewBuildCmdManually().
    UseExecutable("go").
    UseArguments("vet", ".")
```

#### a, `.UseExecutable(binary)`
Specifies the executable to run.

```go
cmd.UseExecutable("go")
```

#### b, `.UseArguments(arg...)`
Specifies the command arguments.

```go
cmd.UseArguments("vet", ".")
```

## IV. Build Targets
A build target groups commands together.

Example:
```go
all := yesb.NewBuildTarget().
    UseName("all").
    UseCommands(hello).
    DependsOn(clean)
```

### 1. `.UseName(name)`
Since `yesb` (and Go itself) have no idea what the variable name you assiged the build target to, it's required for you to manually specify the target name.

It is recommended to set this the same as your variable name to avoid confusion.

```go
target.UseName("target")
```

### 2. `.UseCommands(cmd...)`
Adds commands to the target.

Multiple commands are allowed.

```go
target.UseCommands(compile, link, run)
```

### 3. `.DependsOn(target...)`
Adds dependencies to the target.

Multiple dependencies are allowed.

```go
target.DependsOn(hello)
```

## V. Builder
A builder is responsible for traversing and executing the build graph.

Example:
```go
builder := yesb.NewBuilder().
    UseTargets(clean, all).
    Build()
```

### 1. `.UseTargets(target...)`
Adds targets to the build graph.

Multiple targets are allowed.

```go
builder.UseTargets(program)
```

### 2. `.Build()`
Traverses the build graph usng DFS and executes the required commands.

Dependency cycle are also detected. If a cycle is detected, it will reports the cycle and exits without recursively traversing forever.

It also checks whether each command is up-to-date, and skip them to save time.

It is intended to only called once at the end of `build.go`, or called last on a chained expression (since it returns nothing).
