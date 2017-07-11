package main

import (
	"fmt"
	"os"

	campfire "mixrad.io/klink/campfire"
	common "mixrad.io/klink/common"
	complete "mixrad.io/klink/complete"
	console "mixrad.io/klink/console"
	baker "mixrad.io/klink/baker"
	doctor "mixrad.io/klink/doctor"
	maestro "mixrad.io/klink/maestro"
	flags "mixrad.io/klink/flags"
	git "mixrad.io/klink/git"
	jenkins "mixrad.io/klink/jenkins"
	lister "mixrad.io/klink/lister"
	props "mixrad.io/klink/props"
	pedant "mixrad.io/klink/pedant"
	ssh "mixrad.io/klink/ssh"
	update "mixrad.io/klink/update"
)

func handleAction(args common.Command) {
	// global error handling
	defer func() {
		if p := recover(); p != nil {
			if args.Debug == true {
				console.Red()
				fmt.Println("\nDon't worry about the paths in trace, that's just go.\n")
				console.Reset()
				panic(p)
			}
			console.Red()
			fmt.Println(p)
			console.Reset()
			console.Fail("An error has occured. You may get more information using --debug true")
		}
	}()

	// everything else
	for i := range common.Components {
		component := common.Components[i]
		if args.Action == component.Command {
			component.Callback(args)
			return
		}
	}

	// failed to find the command, print help
	flags.PrintHelpAndExit()
}

func init() {
	// This whole thing makes me sad. Go demands that stuff like this is explicit
	// if we don't reference the namespace then even the .init() function won't be
	// called. We can't reference the namespace without using it so we basically
	// need to manually call the psuedo init methods, Init(), on each component
	// namesapce. Go doesn't allow, or encourage, this kind of aspecty metaprogramming
	campfire.Init()
	complete.Init()
	baker.Init()
	doctor.Init()
	maestro.Init()
	git.Init()
	jenkins.Init()
	lister.Init()
	pedant.Init()
	ssh.Init()
	update.Init()
}

func main() {
	props.EnsureRCFile()
	update.EnsureUpdatedRecently(os.Args[0])
	handleAction(flags.LoadFlags())
}
