package lister

import (
	"encoding/json"
	"fmt"
	"strings"

	jsonq "github.com/jmoiron/jsonq"
	common "mixrad.io/klink/common"
	conf "mixrad.io/klink/conf"
	console "mixrad.io/klink/console"
	props "mixrad.io/klink/props"
)

func Init() {
	common.Register(
		common.Component{"register-app-lister", CreateApp,
			"{app} Creates a new application in lister only", "APPS"},
		common.Component{"info", Info,
			"{app} Return information about the application", "APPS"},
		common.Component{"add-lister-prop", AddProperty,
			"{app} {name} {value} Adds a lister property (json)", "APPS|PROPNAMES"},
		common.Component{"get-lister-prop", GetPropertyFromArgs,
			"{app} {property-name} get the property for the application", "APPS"},
		common.Component{"status", Status,
			"{app} Checks the status of the app", "APPS"},
		common.Component{"apps", ListApps,
			"Lists the applications that exist (via maestro)", "APPS"},
		common.Component{"delete-lister-prop", DeleteProperty,
			"{app} {property-name} Delete the property for the application", "APPS"})
}

type App struct {
	Name string `json:"name"`
}

func listerUrl(end string) string {
	return conf.ListerUrl + end
}

// Return the list of apps that are known about by lister
func GetApps() []string {
	apps, err := common.GetAsJsonq(listerUrl("/applications")).ArrayOfStrings("applications")
	if err != nil {
		panic(err)
	}
	return apps
}

func GetCommonPropertyNames() []string {
	return []string{"bakeType", "baker", "customBakeCommands", "jobsPath", "releasePath", "srcRepo", "servicePathPoke", "statusPath", "testPath"}
}

// List the apps known about by lister
func ListApps(args common.Command) {
	fmt.Println(strings.Join(GetApps(), "\n"))
}

// Create a new application in lister
func CreateApp(args common.Command) {
	if args.SecondPos == "" {
		console.Fail("Must supply an application name as second positional argument")
	}

	createBody := App{args.SecondPos}

	response := common.PostJson(listerUrl("/applications"), createBody)

	fmt.Println("Lister has created our application for us!")
	fmt.Println(response)
}

// Returns true if the app exists
func AppExists(appName string) bool {
	return common.Head(listerUrl("/applications/" + appName))
}

// Returns all information stored in lister about the supplied application
func Info(args common.Command) {
	console.MaybeJQS(common.GetString(listerUrl("/applications/" + args.SecondPos)))
}

func ToJsonValue(in string) (string, error) {
	in = "{\"value\" : " + in + "}"
	var generic interface{}
	err := json.Unmarshal([]byte(in), &generic)
	if err != nil {
		return "", err
	}

	x, err := json.Marshal(generic)
	return string(x), err
}

func AddProperty(args common.Command) {
	app := args.SecondPos
	if app == "" {
		console.Fail("Must supply application name as a second positional argument")
	}
	name := args.ThirdPos
	if name == "" {
		console.Fail("Must supply the property name as the third positional argument")
	}
	value := args.FourthPos
	if value == "" {
		console.Fail("Must supply the property value as the fourth positional argument")
	}

	valueString, err := ToJsonValue(value)
	if err != nil {
		valueString, err = ToJsonValue("\"" + value + "\"")
		if err != nil {
			fmt.Println("That doesn't look like json")
			panic(err)
		}
	}

	fmt.Println(common.PutString(listerUrl("/applications/"+app+"/"+name),
		valueString))
}

func EnsureProp(jq *jsonq.JsonQuery, app string, name string) string {
	str, err := jq.String("metadata", name)
	if err != nil {
		obj, err := jq.Interface("metadata", name)
		if err != nil {
			fmt.Printf(
				"Application %s doesn't have a %s defined, add one with:\n",
				app,
				name,
			)
			console.Fail(fmt.Sprintf("klink add-lister-prop %s %s 'value'\n",
				app, name))
		}
		// this is the only way to get a string from an arbitary type in go...
		return fmt.Sprintf("%s", obj)
	}
	return str
}

func Status(args common.Command) {
	app := args.SecondPos
	if app == "" {
		console.Fail("Must supply application name as a second positional argument")
	}

	jq := common.GetAsJsonq(listerUrl("/applications/" + app))

	statusUrl := EnsureProp(jq, app, "servicePathPoke") + EnsureProp(jq, app, "statusPath")
	fmt.Printf("Checking status at: %s\n", statusUrl)

	console.Green()
	fmt.Println(common.GetString(statusUrl))
	console.Reset()
}

func GetProperty(app string, name string) string {
	jq := common.GetAsJsonq(listerUrl("/applications/" + app))
	return EnsureProp(jq, app, name)
}

func GetOptionalProperty(app string, name string) string {
	jq := common.GetAsJsonq(listerUrl("/applications/" + app))
	str, err := jq.String("metadata", name)
	if err != nil {
		obj, err := jq.Interface("metadata", name)
		if err != nil {
			return ""
		}
		// this is the only way to get a string from an arbitary type in go...
		return fmt.Sprintf("%s", obj)
	}
	return str
}

func GetPropertyFromArgs(args common.Command) {
	app := args.SecondPos
	name := args.ThirdPos
	if app == "" {
		console.Fail("Don't forget to bring a towel^H^H^H^H^H^H pass a application name")
	}
	if name == "" {
		console.Fail("You forgot to pass the property name")
	}
	fmt.Println(GetProperty(app, name))
}

func DeleteProperty(args common.Command) {
	app := args.SecondPos
	name := args.ThirdPos

	if app == "" {
		console.Fail("You forgot to pass the app name")
	}
	if name == "" {
		console.Fail("You forgot to pass the property name")
	}

	common.Delete(listerUrl("/applications/" + app + "/" + name))

	console.Green()
	fmt.Println("Success!")
	console.Reset()
}

// Get the list of environments from lister
func EnvironmentsFromLister() []string {
	envs, err := common.GetAsJsonq(listerUrl("/environments")).ArrayOfStrings("environments")
	if err != nil {
		panic("Unable to parse response getting environments :-(")
	}
	return envs
}

// Returns a list of available environments, accepts an environment
// if that environment isn't known then go and ge the list from
// lister
func GetEnvironments(env string) []string {
	environments := props.GetEnvironments()
	if !common.Contains(environments, env) {
		environments = EnvironmentsFromLister()
		props.SetEnvironments(environments)
	}
	return environments
}

// Returns true if the environment is known by lister
func KnownEnvironment(env string) bool {
	return common.Contains(GetEnvironments(env), env)
}
