package git

import (
	"fmt"
	"os/exec"

	common "mixrad.io/klink/common"
	console "mixrad.io/klink/console"
	lister "mixrad.io/klink/lister"
)

func Init() {
	common.Register(
		common.Component{"clone-tyr", CloneTyrant,
			"{app} {env - optional} clone the tyrant properties for an app into pwd", "APPS"},
		common.Component{"clone-pedant", ClonePedant,
			"{app} {env - optional} clone the pedant properties for an app into pwd", "APPS|ENVS"},
		common.Component{"clone", CloneService,
			"{app} clone the application into pwd", "APPS"},
		common.Component{"gist", Gist,
			"{file-name} [{description}] send stdin to a github gist, use extension to set type", ""})
}

func appName(args common.Command) string {
	if args.SecondPos == "" {
		console.Fail("Application must be provided as the second positional argument")
	}
	return args.SecondPos
}

func envName(args common.Command) string {
	if args.ThirdPos == "" {
		return "all"
	}
	return args.ThirdPos
}

func gitUrlTyrant(app string, env string) string {
	return fmt.Sprintf("git@github.brislabs.com:tyranitar/%s-%s.git", app, env)
}

func gitUrlPedant(app string) string {
	return fmt.Sprintf("git@github.brislabs.com:shuppet/%s.git", app)
}

func gitClone(path string) {
	out, err := exec.Command("git", "clone", path).Output()
	if err != nil {
		fmt.Println(fmt.Sprintf("Error cloning repo, %s, does it already exist? %s", path, err))
	}
	fmt.Println(string(out))
}

// Clone the tyrant properties for the supplied app
func CloneTyrant(args common.Command) {
	app := appName(args)
	env := envName(args)

	if env == "all" {
		gitClone(gitUrlTyrant(app, "poke"))
		gitClone(gitUrlTyrant(app, "prod"))
	} else {
		gitClone(gitUrlTyrant(app, env))
	}
}

// Clone the pedant properties for the supplied app
func ClonePedant(args common.Command) {
	app := appName(args)
	gitClone(gitUrlPedant(app))
}

func CloneService(args common.Command) {
	app := args.SecondPos

	path := lister.GetProperty(app, "srcRepo")
	gitClone(path)
}
