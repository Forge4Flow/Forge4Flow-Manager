package cmd

import (
	"fmt"

	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

// WelcomeMessage to introduce ofc-bootstrap
const WelcomeMessage = "Welcome to Forge4Flow Manager"

func init() {
	rootCommand.AddCommand(versionCmd)
	rootCommand.AddCommand(upCmd)
	rootCommand.AddCommand(installCmd)
	rootCommand.AddCommand(makeProviderCmd())
	rootCommand.AddCommand(collectCmd)
}

func RootCommand() *cobra.Command {
	return rootCommand
}

var (
	// GitCommit Git Commit SHA
	GitCommit string
	// Version version of the CLI
	Version string
)

// Execute faasd
func Execute(version, gitCommit string) error {

	// Get Version and GitCommit values from main.go.
	Version = version
	GitCommit = gitCommit

	if err := rootCommand.Execute(); err != nil {
		return err
	}
	return nil
}

var rootCommand = &cobra.Command{
	Use:   "f4f-manager",
	Short: "Start Forge4Flow-Manager",
	Long: `
Forge4Flow-Manager - Serverless For Everyone Else
`,
	RunE:         runRootCommand,
	SilenceUsage: true,
}

func runRootCommand(cmd *cobra.Command, args []string) error {

	printLogo()
	cmd.Help()

	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information.",
	Run:   parseBaseCommand,
}

func parseBaseCommand(_ *cobra.Command, _ []string) {
	printLogo()

	printVersion()
}

func printVersion() {
	fmt.Printf("f4f-manager version: %s\tcommit: %s\n", GetVersion(), GitCommit)
}

func printLogo() {
	logoText := aec.WhiteF.Apply(Logo)
	fmt.Println(logoText)
}

// GetVersion get latest version
func GetVersion() string {
	if len(Version) == 0 {
		return "dev"
	}
	return Version
}

// Logo for version and root command
const Logo = ` _______  _______  _______  _______  _______   ___    _______  _        _______                  _______  _______  _        _______  _______  _______  _______ 
(  ____ \(  ___  )(  ____ )(  ____ \(  ____ \ /   )  (  ____ \( \      (  ___  )|\     /|       (       )(  ___  )( (    /|(  ___  )(  ____ \(  ____ \(  ____ )
| (    \/| (   ) || (    )|| (    \/| (    \// /) |  | (    \/| (      | (   ) || )   ( |       | () () || (   ) ||  \  ( || (   ) || (    \/| (    \/| (    )|
| (__    | |   | || (____)|| |      | (__   / (_) (_ | (__    | |      | |   | || | _ | | _____ | || || || (___) ||   \ | || (___) || |      | (__    | (____)|
|  __)   | |   | ||     __)| | ____ |  __) (____   _)|  __)   | |      | |   | || |( )| |(_____)| |(_)| ||  ___  || (\ \) ||  ___  || | ____ |  __)   |     __)
| (      | |   | || (\ (   | | \_  )| (         ) (  | (      | |      | |   | || || || |       | |   | || (   ) || | \   || (   ) || | \_  )| (      | (\ (   
| )      | (___) || ) \ \__| (___) || (____/\   | |  | )      | (____/\| (___) || () () |       | )   ( || )   ( || )  \  || )   ( || (___) || (____/\| ) \ \__
|/       (_______)|/   \__/(_______)(_______/   (_)  |/       (_______/(_______)(_______)       |/     \||/     \||/    )_)|/     \|(_______)(_______/|/   \__/
                                                                                                                                                               
`
