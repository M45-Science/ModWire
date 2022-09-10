package utils

import (
	"fmt"
	"strconv"
	"time"

	"../../glob"
	"../../support"
	"github.com/bwmarrin/discordgo"
)

func Stats(s *discordgo.Session, m *discordgo.MessageCreate, args []string, channel string) {
	now := time.Now()
	buffer := ""

	buffer = buffer + fmt.Sprintf("```")

	buffer = buffer + fmt.Sprintf("ModWire Version: %s\n\n", glob.Version)
	buffer = buffer + fmt.Sprintf("Guilds:\n")

	//--- RLOCK ---
	glob.GuildLock.RLock()
	//--- RLOCK ---
	for x := 0; x < glob.GuildDatabaseLen; x++ {
		if !glob.GuildDatabase[x].Deleted {
			buffer = buffer + fmt.Sprintf("ID: %s\nName: %s\nMembers: %d\n\n",
				glob.GuildDatabase[x].ID,
				glob.GuildDatabase[x].Name,
				glob.GuildDatabase[x].MemberCount)

			// Get channels for this guild
			channels, _ := glob.DS.GuildChannels(glob.GuildDatabase[x].ID)

			buffer = buffer + fmt.Sprintf("Channels: ")
			for _, c := range channels {
				// Check if channel is a guild text channel and not a voice or DM channel
				if c.Type != discordgo.ChannelTypeGuildText {
					continue
				}

				// Send text message
				buffer = buffer + fmt.Sprintf("%s, %d", c.Name, c.Position)
			}
			buffer = buffer + fmt.Sprintf("\n\n")

			// Get roles for guild
			buffer = buffer + fmt.Sprintf("Roles: ")

			rlen := glob.GuildDatabase[x].RolesLength
			for rpos := 0; rpos < rlen; rpos++ {
				buffer = buffer + fmt.Sprintf("%s, ", glob.GuildDatabase[x].Roles[rpos].Name)
			}
			buffer = buffer + fmt.Sprintf("\n\n")
		}
	}
	//--- RUNLOCK ---
	glob.GuildLock.RUnlock()
	//--- RUNLOCK ---

	//Sizes
	buffer = buffer + fmt.Sprintf("Stats:\n")
	buffer = buffer + fmt.Sprintf("FactModDB size: %d -- Max: %d\n", glob.FactModDBLen, glob.FactModMax)

	//Search
	rate, _ := strconv.Atoi(support.Config.FactModRefreshRate)
	next := time.Duration(rate) * time.Minute

	diff := now.Sub(glob.LastFactModRefresh)
	till := (next - diff)
	buffer = buffer + fmt.Sprintf("Search Refresh: %s ago. -- Next: %s\n", diff.Round(time.Second), till.Round(time.Second))

	//Uptime
	diff = now.Sub(glob.BootTime)
	buffer = buffer + fmt.Sprintf("Uptime: %s\n", diff.Round(time.Second))

	buffer = buffer + fmt.Sprintf("```")

	_, err := s.ChannelMessageSend(channel, buffer)
	if err != nil {
		support.Log("Couldn't send message: stat command")
	}
}
