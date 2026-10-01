package main

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

    hello := NewBuildTarget()
    hello.UseName("hello")
    hello.UseCommands(compile)

    all := NewBuildTarget()
    all.UseName("all")
    all.UseCommands(run)
    all.DependsOn(hello)

    builder := NewBuilder()
    builder.UseTargets(all)
    builder.Build()
}
