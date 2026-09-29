package main

func main() {
    GoRebuildUrself()

    a := NewBuildTarget()
    a.UseName("a")

    b := NewBuildTarget()
    b.UseName("b")

    c := NewBuildTarget()
    c.UseName("c")

    // a -> b -> c -> a
    a.DependsOn(b)
    b.DependsOn(c)
    c.DependsOn(a)

    builder := NewBuilder()
    builder.UseTargets(a)

    builder.Build()
}
