package main

func main() {
    cpp := NewBuildCmd()
    cpp.UseExecutable("g++")
    cpp.WillUse("hello.cpp")
    cpp.WillOutput("hello")
    cpp.OutputFlag("-o")

    run := NewBuildCmdManually()
    run.UseExecutable("./hello")

    compile := NewBuildTarget()
    compile.UseName("compile")
    compile.UseCommands(cpp)

    exec := NewBuildTarget()
    exec.UseName("exec")
    exec.UseCommands(run)
    exec.DependsOn(compile)

    builder := NewBuilder()
    builder.UseTargets(exec)
    builder.Build()
}
