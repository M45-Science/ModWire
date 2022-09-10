package utils

import (
	"fmt"

	"../../glob"
	"../../support"
	"github.com/bwmarrin/discordgo"
)

func Help(s *discordgo.Session, m *discordgo.MessageCreate, args []string, channel string) {
	buffer := ""
	isadmin := ""

	for _, command := range glob.CL.CommandList {
		if command.Admin {
			isadmin = "(Admin Only) "
		} else {
			isadmin = ""
		}

		if command.Admin && !CheckAdmin(m.Author.ID) {
			//Not an admin, don't display this command.
			continue
		}

		helpb := ""
		if command.HelpB != "" {
			helpb = command.HelpB + "\n\n"
		}

		buffer = buffer + fmt.Sprintf("%s`%s%s %s`\n%s", isadmin, support.Config.DiscordCommandPrefix, command.Name, command.Help, helpb)
	}
	_, err := s.ChannelMessageSend(channel, buffer)
	if err != nil {
		support.Log("Couldn't send message: help command")
	}
	return
}

// CheckAdmin checks if the user attempting to run an admin command is an admin
func CheckAdmin(ID string) bool {
	for _, admin := range support.Config.DiscordAdminIDs {
		if ID == admin {
			return true
		}
	}
	return false
}
