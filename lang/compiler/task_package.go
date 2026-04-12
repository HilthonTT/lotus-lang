package compiler

import (
	"fmt"
	"sync"
	"time"

	"github.com/hilthontt/lotus/object"
)

// taskState holds a shared WaitGroup so Task.wait() works correctly.
type taskState struct {
	wg sync.WaitGroup
	mu sync.Mutex
}

var globalTaskState = &taskState{}

func taskPackage() *object.Package {
	pkg := &object.Package{
		Name:      "Task",
		Functions: map[string]object.PackageFunction{},
	}

	state := &taskState{}

	// Task.spawn(fn) - Runs a Lotus closure in a new goroutine
	pkg.Functions["spawn"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		closure, ok := args[0].(*object.Closure)
		if !ok {
			return &object.Nil{}
		}

		state.wg.Add(1)
		go func() {
			defer state.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Task panic: %v\n", r)
				}
			}()

			if pkg.CallVM != nil {
				pkg.CallVM(closure, []object.Object{})
			}
		}()

		return &object.Nil{}
	}

	// Task.spawnWith(fn, arg) — spawn with a single argument
	pkg.Functions["spawnWith"] = func(args ...object.Object) object.Object {
		if len(args) < 2 {
			return &object.Nil{}
		}
		closure, ok := args[0].(*object.Closure)
		if !ok {
			return &object.Nil{}
		}
		fnArg := args[1]

		state.wg.Add(1)
		go func() {
			defer state.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Task panic: %v\n", r)
				}
			}()

			if pkg.CallVM != nil {
				pkg.CallVM(closure, []object.Object{fnArg})
			}
		}()

		return &object.Nil{}
	}

	// Task.wait() — block until all spawned tasks finish
	pkg.Functions["wait"] = func(args ...object.Object) object.Object {
		state.wg.Wait()
		return &object.Nil{}
	}

	pkg.Functions["sleep"] = func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Nil{}
		}
		ms, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Nil{}
		}
		time.Sleep(time.Duration(ms.Value) * time.Millisecond)
		return &object.Nil{}
	}

	// Task.mutex() — returns a simple lock object
	pkg.Functions["mutex"] = func(args ...object.Object) object.Object {
		m := &object.Mutex{}
		return m
	}

	return pkg
}
