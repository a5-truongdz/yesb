# yesb - YesBuild

`yesb` is a small build system written in Go, inspired by Tsoding's [nob.h](https://github.com/tsoding/nob.h) idea.

The name stands for "YesBuild", because it's basically the opposite of NoBuild

The idea is simple: instead of describing your build using another build language or configuration format, you write your build system directly in Go.

## NoBuild Idea

The idea is that you should not need a separate build tool or build language to build your project.

Instead of:

- Make
- CMake
- Ninja
- shell scripts
- complicated configuration files

you can use the programming language you already know to describe how your project should be built.

With `yesb`, that language is Go.

Your build system is just a Go program.

It can execute commands, inspect files, track dependencies, rebuild things when necessary, and do whatever else you need — because you have the full power of Go available.

## This is an Experimental Project

`yesb` is an experimental project and is still under development.

I'm building it mostly to explore how a build system works internally rather than to replace established tools.

The implementation is intentionally small and straightforward. Things like dependency tracking, incremental builds, command execution, and dependency graphs are implemented directly instead of being hidden behind a large framework.

The API may change at any time.

## It's likely Not Suitable for Your Project

If your project already uses a mature build system and it works well for you, there is probably little reason to replace it with `yesb`.

`yesb` is more suitable for small projects, personal projects, experiments, and people who would rather write a build program than learn another build configuration language.

It is also probably not a great choice if your build requires a huge amount of platform-specific logic or dependency discovery.

In that case, you may want an established build system instead.

## Advantages

- Your build system is written in Go.
- No separate build language is required.
- You can use the full Go standard library.
- Build logic can be expressed directly using normal Go code.
- You can implement custom build commands when the built-in abstractions are not enough.
- The implementation is small enough to understand and modify yourself.
- You get to use Go more.
- ...

## Disadvantages

- You need to know Go.
- You need to implement things yourself.
- The project is experimental.
- It does not try to provide every feature of mature build systems.
- Portability depends on the commands and tools used by your build script.
- You get to use Go more.
- You need Go.
- You possess Go.
- You love Go.
- ...

## How to Use

The basic idea is that your build script is simply a Go program.

- Run `go get github.com/a5-truongdz/yesb`

- Create a `build.go` and define your build targets and their dependencies,

- Bootstrap it with `go build -tags build build.go`

After that you can just call `./build` to make it auto-rebuild.

For the current syntax and examples, please wait. It's finalizing.

There is intentionally no separate configuration syntax to learn.

If you know Go, you already know the language used to describe your build.

## Dependency Graph

Build targets form a directed dependency graph.

For example:

```text
        program
        /     \
    compile   assets
       |
   source.cpp
````

When building `program`, `yesb` traverses its dependencies first and then builds the target itself.

Shared dependencies are only built once.

Cycles are detected as well:

```text
A → B → C
    ↑   ↓
    └───┘
```

A dependency cycle results in a build error instead of recursively traversing forever.

## Incremental Builds

`yesb` can determine whether a command needs to run again by comparing the modification time of its inputs and output.

If `program` already exists and is newer than `source.cpp`, the command can be skipped.

If `source.cpp` is newer, the command is executed again.

This keeps repeated builds fast without requiring a separate dependency database.

## Why Go?

Go is a convenient language for this approach because it has:

* simple process execution through `os/exec`
* filesystem APIs through `os` and `io/fs`
* straightforward concurrency primitives
* easy cross-compilation
* a small standard library
* a simple language that is suitable for writing small tools

Most importantly, the build system itself is just a normal Go program.

## Inspiration

`yesb` is heavily inspired by the NoBuild approach and especially by Tsoding's [`nob.h`](https://github.com/tsoding/nob.h).

This README this heavily inspired too.

The goal is not to directly reproduce `nob.h`, but to explore the same general idea in Go.

## Status

This project is experimental.

Expect bugs, missing features, questionable API decisions, and things that will probably be rewritten later.

That's part of the point.

---

*thanks, tsoding*
