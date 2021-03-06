package complete

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	common "mixrad.io/klink/common"
	console "mixrad.io/klink/console"
	baker "mixrad.io/klink/baker"
	lister "mixrad.io/klink/lister"
	props "mixrad.io/klink/props"
)

// register our public command with klink
func Init() {
	common.Register(
		common.Component{"gen-complete", GenComplete, "Refresh the autocomplete data", ""})
}

// returns a path to the file in klink's directory
func filePath(path string) string {
	klinkDir := common.UserHomeDir() + "/.klink.d/"
	if !common.Exists(klinkDir) {
		os.MkdirAll(klinkDir, 0755)
	}
	return klinkDir + path
}

// write the array of strings to the path inside the klink directory
func stringsToFile(path string, contents []string) {
	ioutil.WriteFile(filePath(path), []byte(strings.Join(contents, "\n")), 0644)
}

// generate the environments from lister, passing not found forces a refresh
func generateEnvs() {
	fmt.Println("Generating environment file")
	stringsToFile("/envs", lister.GetEnvironments("notfound"))
}

func generatePropertyNames() {
	fmt.Println("Generating common property names")
	stringsToFile("/propnames", lister.GetCommonPropertyNames())
}

// generate the app list from lister
func generateApps() {
	fmt.Println("Generating app file")
	stringsToFile("/apps", lister.GetApps())
}

// generate a list of campfire rooms
func generateRooms() {
	fmt.Println("Generating rooms file")
	stringsToFile("/rooms", console.Rooms())
}

// autocomplete options just for klink baker. what fun!
func generateBakerHelpers() {
	fmt.Println("Generating baker helpers")
	stringsToFile("/baker", baker.HelperNames())
}

// generate the list of all klink commands
func generateCommands() {
	fmt.Println("Generating command file")
	stringsToFile("/commands", common.ComponentNames())
}

// generate the list of arg types that each command takes
func generateCommandArgs() {
	fmt.Println("Generating command args file")

	var acs []string
	acs = append(acs, "COMMANDFORMATS=( ")
	for _, component := range common.Components {
		acs = append(acs, fmt.Sprintf("\"%s:%s\"",
			component.Command, component.AutoComplete))
	}
	acs = append(acs, " )")
	stringsToFile("/command_ac_formats", acs)
}

// generate the autocomplete script
func generateScript() {
	fmt.Println("Generating the auto complete script")

	script := `#!/bin/bash

kpath="$HOME/.klink.d"

source $kpath/command_ac_formats

function mget {
    for animal in "${COMMANDFORMATS[@]}" ; do
        KEY="${animal%%:*}"
        VALUE="${animal##*:}"
        if [ $KEY = $1 ]; then
            MVAL=$VALUE
        fi
    done
}

function get_complete {
    case $1 in
        "APPS")
            COMPREPLY=($(compgen -W "$(cat $kpath/apps)" -- $cur))
            ;;
        "ENVS")
            COMPREPLY=($(compgen -W "$(cat $kpath/envs)" -- $cur))
            ;;
        "ROOMS")
            COMPREPLY=($(compgen -W "$(cat $kpath/rooms)" -- $cur))
            ;;
        "BAKERS")
            COMPREPLY=($(compgen -W "$(cat $kpath/bakers)" -- $cur))
            ;;
        "PROPNAMES")
            COMPREPLY=($(compgen -W "$(cat $kpath/propnames)" -- $cur))
            ;;
        "_")
            COMPREPLY=()
            ;;
        *)
            COMPREPLY=()
            ;;
    esac
}

_klink()
{
	local cur=${COMP_WORDS[COMP_CWORD]}

	case ${COMP_CWORD} in
		1)
            COMPREPLY=($(compgen -W "$(cat $kpath/commands)" -- $cur))
			;;
		*)
            local top=${COMP_WORDS[1]}
            mget $top
            local command_string=$MVAL
            local command_list=(${command_string//|/ })
            get_complete ${command_list[$COMP_CWORD-2]}
			;;
	esac
}
complete -F _klink klink`

	ioutil.WriteFile(filePath("/klink_autocomplete.bash"),
		[]byte(script), 0644)
}

// ensure that a command to source the autocomplete script is written to bash
func addSourceToBash() {
	if !props.HasAutoCompleteRun() {
		homeFile := ""
		homeFile = common.UserHomeDir() + "/.bashrc"
		if common.Exists(homeFile) == false {
			homeFile = common.UserHomeDir() + "/.bash_profile"
			if !(common.Exists(homeFile)) {
				return
			}
		}

		console.Green()
		fmt.Println("Adding the source command to: " + homeFile)
		console.Reset()

		f, err := os.OpenFile(homeFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		sourceScript := `
# Added by klink, sets up autocomplete
source $HOME/.klink.d/klink_autocomplete.bash
`
		if _, err = f.WriteString(sourceScript); err != nil {
			panic(err)
		}

		props.SetAutoCompleteHasRun()
	}
}

// generate everything required for klink autocomplete to work
func GenComplete(_ common.Command) {
	generateEnvs()
	generateApps()
	generatePropertyNames()
	generateRooms()
	generateBakerHelpers()
	generateCommands()
	generateCommandArgs()
	generateScript()
	addSourceToBash()
}
