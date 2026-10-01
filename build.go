package main

import (
    "os"
    "log"
)

func main() {
    GoRebuildUrself()

    compile := NewBuildCmd()
    compile.UseExecutable("g++")
    compile.WillUse("hello.cpp")
    compile.WillOutput("hello")
    compile.OutputFlag("-o")
    compile.UseFlags(
        "-std=c++20",
        "-D_GLIBCXX_DEBUG",
        "-Wall", "-Wextra",
        "-Wpedantic",
        "-g3", "-v",
    )

    run := NewBuildCmdManually()
    run.UseExecutable("./hello")

    all := NewBuildTarget()
    all.UseName("all")
    all.UseCommands(compile, run)

    del := NewBuildCmdManually()
    del.UseExecutable("rm")
    del.UseArguments("hello")

    clean := NewBuildTarget()
    clean.UseName("clean")
    clean.UseCommands(del)

    builder := NewBuilder()

    if len(os.Args) == 1 {    // only `./build`
        builder.UseTargets(all)
    } else {
        if os.Args[1] == "clean" {
            builder.UseTargets(clean)
        } else if os.Args[1] == "all" {
            builder.UseTargets(all)
        } else {
            log.Fatalf("unknown target %s", os.Args[1])
        }
    }

    builder.Build()
}
