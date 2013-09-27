package main

import (
	"flag"
	"fmt"
	args "nokia.com/klink/args"
	console "nokia.com/klink/console"
	ditto "nokia.com/klink/ditto"
	exploud "nokia.com/klink/exploud"
	onix "nokia.com/klink/onix"
)

func loadFlags() args.Command {
	command := args.Command{}
	flag.StringVar(&command.Action, "action", "", "Action for klink to perform: [deploy, build, rollback]")
	flag.StringVar(&command.Application, "app", "", "Application to do stuff with")
	flag.StringVar(&command.Ami, "ami", "deploy", "Set the ami to deploy")
	flag.StringVar(&command.Environment, "env", "ent-dev", "Environment for this command [ent-dev, prod]")
	flag.StringVar(&command.Message, "m", "", "Message for this action. For example why you are deploying")
	flag.StringVar(&command.Version, "v", "", "Version to deploy, rollback etc [e.g. 0.153]")
	flag.Parse()

	fmt.Println("\tAction:      ", command.Action)
	fmt.Println("\tAmi:         ", command.Ami)
	fmt.Println("\tApplication: ", command.Application)
	fmt.Println("\tEnvironment: ", command.Environment)
	fmt.Println("\tMessage:     ", command.Message)
	fmt.Println("\tVersion:     ", command.Version, "\n")
	return command
}

func handleAction(args args.Command) {
	switch args.Action {
	case "deploy":
		exploud.Exploud(args)
	case "bake":
		ditto.Bake(args)
	case "create-service":
		onix.CreateService(args)
    case "list-services":
        onix.ListServices()
	default:
		console.Fail(fmt.Sprintf("Unknown or no action: %s", args.Action))
	}
}

func main() {
	console.Klink()
	command := loadFlags()
	handleAction(command)
}
