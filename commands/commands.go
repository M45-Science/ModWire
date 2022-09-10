package commands

import (
	"strings"

	"../glob"
	"./utils"
	"github.com/bwmarrin/discordgo"
)

// CL is a Commands interface.

// RegisterCommands registers the commands on start up.
func RegisterCommands() {
	// Admin Commands
	//glob.CL.CommandList = append(glob.CL.CommandList, glob.Command{Name: "addmod", Command: admin.Addmod, Admin: true, Help: "<modname>", HelpB: "Adds mod called <modname> to watch list"})
	//glob.CL.CommandList = append(glob.CL.CommandList, glob.Command{Name: "delmod", Command: admin.Delmod, Admin: true, Help: "<modname>", HelpB: "Removes mod called <modname> from watch list"})

	// User Commands
	//glob.CL.CommandList = append(glob.CL.CommandList, glob.Command{Name: "list", Command: utils.Listmods, Admin: false, Help: "", HelpB: "(Shows all tracked mods.)"})
	glob.CL.CommandList = append(glob.CL.CommandList, glob.Command{Name: "info", Command: utils.ModInfo, Admin: false, Help: "<modname>", HelpB: "Shows last update for that mod."})
	glob.CL.CommandList = append(glob.CL.CommandList, glob.Command{Name: "stat", Command: utils.Stats, Admin: false, Help: "", HelpB: "(Shows bot stats)"})

	glob.CL.CommandList = append(glob.CL.CommandList, glob.Command{Name: "help", Command: utils.Help, Admin: false, Help: "(You are here)",
		HelpB: "\n`If you don't know the exact name of the mod,\ntype a keyword from the name and I'll show a list of possible matches.\nAlternatively, you can copy/paste the URL from the mod portal and I can extract the mod name from it.`\n"})

}

// RunCommand runs a specified command.
func RunCommand(name string, s *discordgo.Session, m *discordgo.MessageCreate, args []string, channel string) {
	for _, command := range glob.CL.CommandList {
		if strings.ToLower(command.Name) == strings.ToLower(name) {
			if command.Admin && utils.CheckAdmin(m.Author.ID) {
				command.Command(s, m, args, channel)
			}

			if command.Admin == false {
				command.Command(s, m, args, channel)
			}
			return
		}
	}
}
