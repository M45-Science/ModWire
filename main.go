package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"./commands"
	"./database"
	"./expires"
	"./factorio"
	"./glob"
	"./support"
	"github.com/bwmarrin/discordgo"
)

//-------------------------
//-- FUNCTIONS --
//-------------------------
//Main, short and sweet
func main() {

	now := time.Now()
	glob.LastFactModRefresh = now
	glob.BootTime = now

	support.Log("Hello!")
	glob.Running = true

	support.Config.LoadEnv()
	expires.ReadExpireFile()

	botsetup()

	database.LoadFMDB()
	database.LoadGuilds()

	makegdb := false
	if glob.GuildDatabaseLen == 0 {
		makegdb = true
		support.Log("GDB not found, building...")
	}
	updateguilds(false)
	if makegdb {
		database.SaveGuilds()
	}

	factorio.Get_Username_Token()

	//Fetch search if it is missing
	if glob.FactModDBLen <= 0 {
		support.Log("FMDB not found, fetching...")
		glob.LastFactModRefresh = time.Now()
		factorio.FetchFactMods(true)
	}
	commands.RegisterCommands()

	get_guild_members()

	mloop()

	expires.WriteExpireFile()
	support.Log("Goodbye.")
	shutdownchecks()
	os.Exit(1)
}

func get_guild_members() {
	support.Log("Building guild member cache.")
	for x := 0; x < glob.GuildDatabaseLen; x++ {
		glob.DS.RequestGuildMembers(glob.GuildDatabase[x].ID, "", glob.MaxMembers, false)
	}
}

//-------------------------
//-- BOT SETUP --
//-------------------------
func botsetup() {
	support.Log("Starting bot...")

	if glob.XDEBUG {
		support.Log("Discord Token: ", support.Config.DiscordToken)

	}
	bot, err := discordgo.New("Bot " + support.Config.DiscordToken)
	glob.DS = bot

	if err != nil {
		support.Log("Error creating Discord session: ", err)
		support.ErrorLog("An error occurred when attempting to create the Discord session", err)
		os.Exit(1)
	}

	err = glob.DS.Open()

	if err != nil {
		support.Log("error opening connection,", err)
		support.ErrorLog("An error occurred when attempting to connect to Discord", err)
		os.Exit(1)
	}
	//Wait a moment, to miss inconsistant starting events
	time.Sleep(2 * time.Second)

	bot.AddHandler(IncomingMessage)

	bot.AddHandler(onGuildUpdate)
	bot.AddHandler(onGuildCreate)
	bot.AddHandler(onGuildDelete)

	bot.AddHandler(onGuildMemberAdd)
	bot.AddHandler(onGuildMemberUpdate)
	bot.AddHandler(onGuildMemberRemove)

	bot.AddHandler(onGuildRoleCreate)
	bot.AddHandler(onGuildRoleUpdate)
	bot.AddHandler(onGuildRoleDelete)

	bot.AddHandler(onChannelCreate)
	bot.AddHandler(onChannelUpdate)
	bot.AddHandler(onChannelDelete)

	bot.AddHandler(onGuildMembersChunk)

	errb := glob.DS.UpdateStreamingStatus(0, "mod updates", support.Config.TwitchURL)
	if errb != nil {
		support.Log("UpdateStreamingStatus failed.")
	}
	support.Log("Bot is now running.  Press CTRL-C to exit.")
}

//--------------------------------
//-- INCOMING MESSAGES, DISCORD --
//--------------------------------
func IncomingMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	go func() {

		if m.Author.ID == s.State.User.ID {
			//Hello, no this is Patrick!
			return
		}

		if m.Author.Bot {
			//Don't listen to other bots...
			return
		}

		channel, _ := s.Channel(m.ChannelID)

		//Right channel?
		if m.ChannelID == support.Config.DiscordCmdChannel || channel.Type == discordgo.ChannelTypeDM {
			//Has prefix?
			if strings.HasPrefix(m.Content, support.Config.DiscordCommandPrefix) {

				command := strings.Split(m.Content[1:len(m.Content)], " ")
				name := strings.ToLower(command[0])

				//Run command
				commands.RunCommand(name, s, m, command, m.ChannelID)
				return
			}
			return
		}
	}()
}

//-------------------------
//-- MAIN LOOP --
//-------------------------
func mloop() {

	//Search database refresh
	go func() {
		for {
			sleepfor := expires.CacheRefreshDelay()
			support.Log("Search refresh sleeping for: ", sleepfor)

			time.Sleep(sleepfor)

			factorio.FetchFactMods(false)
			glob.LastFactModRefresh = time.Now()
		}
	}()

	//ModDB database save loop
	go func() {
		for {
			time.Sleep(10 * time.Second)

			glob.FactModDirtyLock.Lock()
			wasdirty := glob.FactModDirty
			glob.FactModDirty = false
			glob.FactModDirtyLock.Unlock()

			if wasdirty {
				database.SaveFMDB()
				time.Sleep(30 * time.Second)
			}
		}
	}()

	//Guild database save loop
	go func() {
		for {
			time.Sleep(10 * time.Second)

			glob.GuildDirtyLock.Lock()
			wasdirty := glob.GuildDirty
			glob.GuildDirty = false
			glob.GuildDirtyLock.Unlock()

			if wasdirty {
				database.SaveGuilds()
				time.Sleep(30 * time.Second)
			}
		}
	}()

	//Members caching
	go func() {
		time.Sleep(60 * time.Minute)
		get_guild_members()
	}()

	go func() {
		return
		for {
			time.Sleep(10 * time.Second)
			_, err := glob.DS.ChannelMessageSend(support.Config.DiscordCmdChannel, "test")
			if err != nil {
				support.Log("test message failed...")
			}
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	support.Log("I've been killed... rosebud.")
	shutdownchecks()
	glob.DS.Close()
	os.Exit(1)
}

//Guild database updating
func updateguilds(newserver bool) {

	support.Log("Updating guilds...")
	//--- LOCK ---
	glob.GuildLock.Lock()
	defer glob.GuildLock.Unlock()
	//--- DEFER --

	for pos := 0; pos < glob.GuildDatabaseLen; pos++ {
		found := false
		for _, g := range glob.DS.State.Guilds {

			if glob.GuildDatabase[pos].ID == g.ID {
				found = true
				//Already in our list

				glob.GuildDatabase[pos].ID = g.ID
				glob.GuildDatabase[pos].Name = g.Name
				glob.GuildDatabase[pos].Icon = g.Icon
				glob.GuildDatabase[pos].MemberCount = g.MemberCount
				glob.GuildDatabase[pos].OwnerID = g.OwnerID
				glob.GuildDatabase[pos].WidgetChannelID = g.WidgetChannelID
				glob.GuildDatabase[pos].SystemChannelID = g.SystemChannelID

				//Rebuild channels
				glob.GuildDatabase[pos].ChannelsLength = 0

				//Get channels for this guild
				channels, _ := glob.DS.GuildChannels(g.ID)
				for _, c := range channels {
					// Skip if a voice or DM channel
					if c.Type != discordgo.ChannelTypeGuildText {
						continue
					}

					cpos := glob.GuildDatabase[pos].ChannelsLength
					glob.GuildDatabase[pos].Channels[cpos].Name = c.Name
					glob.GuildDatabase[pos].Channels[cpos].ID = c.ID
					glob.GuildDatabase[pos].Channels[cpos].Position = c.Position
					glob.GuildDatabase[pos].Channels[cpos].ParentID = c.ParentID
					glob.GuildDatabase[pos].ChannelsLength++
				}

				//Rebuild roles
				glob.GuildDatabase[pos].RolesLength = 0

				//Get roles for this guild
				rpos := glob.GuildDatabase[pos].RolesLength
				for _, r := range g.Roles {
					glob.GuildDatabase[pos].Roles[rpos].ID = r.ID
					glob.GuildDatabase[pos].Roles[rpos].Name = r.Name
					glob.GuildDatabase[pos].Roles[rpos].Managed = r.Managed
					glob.GuildDatabase[pos].Roles[rpos].Hoist = r.Hoist
					glob.GuildDatabase[pos].Roles[rpos].Color = r.Color
					glob.GuildDatabase[pos].Roles[rpos].Position = r.Position
					glob.GuildDatabase[pos].Roles[rpos].Permissions = r.Permissions

					rpos++
				}
				glob.GuildDatabase[pos].RolesLength = rpos

				support.Log("Updating guild:", g.Name)

				//Mark Dirty
				glob.GuildDirtyLock.Lock()
				glob.GuildDirty = true
				glob.GuildDirtyLock.Unlock()
			}
		}
		if !found {
			//Found in our list, but not current list of guilds (deleted)
			glob.GuildDatabase[pos].Deleted = true
			support.Log("Guild deleted: ", glob.GuildDatabase[pos].Name)

			//Mark Dirty
			glob.GuildDirtyLock.Lock()
			glob.GuildDirty = true
			glob.GuildDirtyLock.Unlock()
		}

	}
	for _, g := range glob.DS.State.Guilds {
		found := false
		for pos := 0; pos < glob.GuildDatabaseLen; pos++ {
			if glob.GuildDatabase[pos].ID == g.ID {
				found = true

				//Already in our list
				if glob.GuildDatabase[pos].Deleted {
					glob.GuildDatabase[pos].Deleted = false
					support.Log("Guild Undeleted: ", glob.GuildDatabase[pos].Name)
					glob.GuildDatabase[pos].ID = g.ID
					glob.GuildDatabase[pos].Name = g.Name
					glob.GuildDatabase[pos].Icon = g.Icon
					glob.GuildDatabase[pos].MemberCount = g.MemberCount
					glob.GuildDatabase[pos].OwnerID = g.OwnerID
					glob.GuildDatabase[pos].WidgetChannelID = g.WidgetChannelID
					glob.GuildDatabase[pos].SystemChannelID = g.SystemChannelID

					spithelp(glob.GuildDatabase[pos])

					//Mark Dirty
					glob.GuildDirtyLock.Lock()
					glob.GuildDirty = true
					glob.GuildDirtyLock.Unlock()
				}
			}
		}
		if !found {
			//Not in our list
			if glob.GuildDatabaseLen < glob.GuildsMax-1 {

				//Guild Main Info
				glob.GuildDatabase[glob.GuildDatabaseLen].ID = g.ID
				glob.GuildDatabase[glob.GuildDatabaseLen].Name = g.Name
				glob.GuildDatabase[glob.GuildDatabaseLen].Icon = g.Icon
				glob.GuildDatabase[glob.GuildDatabaseLen].MemberCount = g.MemberCount
				glob.GuildDatabase[glob.GuildDatabaseLen].OwnerID = g.OwnerID

				glob.GuildDatabase[glob.GuildDatabaseLen].WidgetChannelID = g.WidgetChannelID
				glob.GuildDatabase[glob.GuildDatabaseLen].SystemChannelID = g.SystemChannelID

				//Get channels for this guild
				channels, _ := glob.DS.GuildChannels(g.ID)
				for _, c := range channels {
					// Skip if a voice or DM channel
					if c.Type != discordgo.ChannelTypeGuildText {
						continue
					}

					cpos := glob.GuildDatabase[glob.GuildDatabaseLen].ChannelsLength
					glob.GuildDatabase[glob.GuildDatabaseLen].Channels[cpos].Name = c.Name
					glob.GuildDatabase[glob.GuildDatabaseLen].Channels[cpos].ID = c.ID
					glob.GuildDatabase[glob.GuildDatabaseLen].Channels[cpos].Position = c.Position
					glob.GuildDatabase[glob.GuildDatabaseLen].Channels[cpos].ParentID = c.ParentID

					glob.GuildDatabase[glob.GuildDatabaseLen].ChannelsLength++
				}

				//Get roles for this guild
				glob.GuildDatabase[glob.GuildDatabaseLen].RolesLength = 0
				for _, r := range g.Roles {
					rpos := glob.GuildDatabase[glob.GuildDatabaseLen].RolesLength
					pos := glob.GuildDatabaseLen
					glob.GuildDatabase[pos].Roles[rpos].ID = r.ID
					glob.GuildDatabase[pos].Roles[rpos].Name = r.Name
					glob.GuildDatabase[pos].Roles[rpos].Managed = r.Managed
					glob.GuildDatabase[pos].Roles[rpos].Hoist = r.Hoist
					glob.GuildDatabase[pos].Roles[rpos].Color = r.Color
					glob.GuildDatabase[pos].Roles[rpos].Position = r.Position
					glob.GuildDatabase[pos].Roles[rpos].Permissions = r.Permissions
					glob.GuildDatabase[glob.GuildDatabaseLen].RolesLength++
				}

				glob.GuildDatabaseLen++
				support.Log("Adding new guild:", g.Name)
				spithelp(glob.GuildDatabase[glob.GuildDatabaseLen])

				//Mark Dirty
				glob.GuildDirtyLock.Lock()
				glob.GuildDirty = true
				glob.GuildDirtyLock.Unlock()
			} else {
				support.Log("Couldn't add all guilds, limit reached.")
				break

			}
		}
	}
	//buffer := fmt.Sprintf("mods to %d servers", glob.GuildDatabaseLen)
	tusers := 0
	for p := 0; p < glob.GuildDatabaseLen; p++ {
		tusers = tusers + glob.GuildDatabase[p].MemberCount
	}
	buffer := fmt.Sprintf("mods to %d gamers", tusers)
	err := glob.DS.UpdateStreamingStatus(0, buffer, support.Config.TwitchURL)
	if err != nil {
		support.Log("UpdateStreamingStatus failed.")
	}
}

//-------------------------
//-- HOOKS --
//-------------------------
//Guilds

func onGuildUpdate(s *discordgo.Session, event *discordgo.GuildUpdate) {
	if event.Guild.Unavailable == true {
		return
	}

	support.Log("EVENT: Guild update.")
	updateguilds(false)
}
func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable == true {
		return
	}

	support.Log("EVENT: Guild create.")
	updateguilds(true)

}
func onGuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	if event.Guild.Unavailable == true {
		return
	}

	support.Log("EVENT: Guild delete.")
	updateguilds(false)
}

//Members
func onGuildMemberAdd(s *discordgo.Session, member *discordgo.GuildMemberAdd) {
	avail := ""

	for g := 0; g < glob.GuildDatabaseLen; g++ {
		if glob.GuildDatabase[g].ID == member.GuildID {
			if glob.GuildDatabase[g].Unavailable {
				avail = "Guild unavailable."
				return
			}
		}
	}
	support.Log("EVENT: Member add. ", avail)
	updateguilds(false)
	get_guild_members()
}
func onGuildMemberUpdate(s *discordgo.Session, member *discordgo.GuildMemberUpdate) {
	avail := ""

	for g := 0; g < glob.GuildDatabaseLen; g++ {
		if glob.GuildDatabase[g].ID == member.GuildID {
			if glob.GuildDatabase[g].Unavailable {
				avail = "Guild unavailable."
				return
			}
		}
	}
	support.Log("EVENT: Member update.", avail)
	updateguilds(false)
	get_guild_members()

}
func onGuildMemberRemove(s *discordgo.Session, member *discordgo.GuildMemberRemove) {
	avail := ""

	for g := 0; g < glob.GuildDatabaseLen; g++ {
		if glob.GuildDatabase[g].ID == member.GuildID {
			if glob.GuildDatabase[g].Unavailable {
				avail = "Guild unavailable."
				return
			}
		}
	}
	support.Log("EVENT: Member remove.", avail)
	updateguilds(false)
	get_guild_members()
}

//Roles
func onGuildRoleCreate(s *discordgo.Session, GuildRole *discordgo.GuildRoleCreate) {
	avail := ""

	for g := 0; g < glob.GuildDatabaseLen; g++ {
		if glob.GuildDatabase[g].ID == GuildRole.GuildID {
			if glob.GuildDatabase[g].Unavailable {
				avail = "Guild unavailable."
				return
			}
		}
	}
	support.Log("EVENT: Role create.", avail)
	updateguilds(false)
	get_guild_members()
}
func onGuildRoleUpdate(s *discordgo.Session, GuildRole *discordgo.GuildRoleUpdate) {
	avail := ""

	for g := 0; g < glob.GuildDatabaseLen; g++ {
		if glob.GuildDatabase[g].ID == GuildRole.GuildID {
			if glob.GuildDatabase[g].Unavailable {
				avail = "Guild unavailable."
				return
			}
		}
	}
	support.Log("EVENT: Role update.", avail)
	updateguilds(false)
	get_guild_members()
}
func onGuildRoleDelete(s *discordgo.Session, GuildRole *discordgo.GuildRoleDelete) {
	avail := ""

	for g := 0; g < glob.GuildDatabaseLen; g++ {
		if glob.GuildDatabase[g].ID == GuildRole.GuildID {
			if glob.GuildDatabase[g].Unavailable {
				avail = "Guild unavailable."
				return
			}
		}
	}
	support.Log("EVENT: Role delete.", avail)
	updateguilds(false)
	get_guild_members()
}

//Channels
func onChannelCreate(s *discordgo.Session, Channel *discordgo.ChannelCreate) {
	support.Log("EVENT: Channel create.")
	updateguilds(false)
}
func onChannelUpdate(s *discordgo.Session, Channel *discordgo.ChannelUpdate) {
	support.Log("EVENT: Channel update.")
	updateguilds(false)
}
func onChannelDelete(s *discordgo.Session, Channel *discordgo.ChannelDelete) {
	support.Log("EVENT: Channel delete.")
	updateguilds(false)
}

func onGuildMembersChunk(s *discordgo.Session, gm *discordgo.GuildMembersChunk) {
	glob.GuildMemberChunkLock.Lock()
	defer glob.GuildMemberChunkLock.Unlock()

	guild, err := glob.DS.Guild(gm.GuildID)

	if err == nil {
		support.Log("GMC: ", guild.Name)
		msize := len(gm.Members) - 1

		guildpos := -1
		for gpos := 0; gpos < glob.GuildDatabaseLen; gpos++ {
			if glob.GuildDatabase[gpos].ID == gm.GuildID {
				guildpos = gpos
				break
			}
		}
		if guildpos < 0 {
			support.Log("Couldn't find guild in GDB to match GuildMemberChunk.")
			return
		}

		adminrole := "None found"
		rlen := glob.GuildDatabase[guildpos].RolesLength
		var roles [glob.RolesMax]string
		roleslen := 0

		//Find admin roles on guild
		for rpos := 0; rpos < rlen; rpos++ {
			if glob.GuildDatabase[guildpos].Roles[rpos].Permissions&discordgo.PermissionManageWebhooks == discordgo.PermissionManageWebhooks {
				adminrole = glob.GuildDatabase[guildpos].Roles[rpos].Name
				support.Log(fmt.Sprintf("admin role: guild: %s role: %s", glob.GuildDatabase[guildpos].Name, adminrole))
				roles[rpos] = glob.GuildDatabase[guildpos].Roles[rpos].ID
				roleslen = rpos
			}
		}

		roleslen++

		//Find users that have admin roles
		glob.GuildDatabase[guildpos].AdminsLen = 0
		apos := 0
		for x := 0; x < msize; x++ {
			for cpos, _ := range gm.Members[x].Roles {
				for rpos := 0; rpos < roleslen; rpos++ {
					if gm.Members[x].Roles[cpos] == roles[rpos] {
						support.Log(fmt.Sprintf("Admin: %s %s", glob.GuildDatabase[guildpos].Name, gm.Members[x].User.Username))
						glob.GuildDatabase[guildpos].Admins[apos].ID = gm.Members[x].User.ID
						apos++
						glob.GuildDatabase[guildpos].AdminsLen = apos
					}
				}
			}
		}
		support.Log(fmt.Sprintf("Users added from GMC: %s: %d ", glob.GuildDatabase[guildpos].Name, glob.GuildDatabase[guildpos].AdminsLen))

	} else {
		support.Log("GMC: Couldn't find guild by ID.")
	}
}

func spithelp(guild glob.GuildFormat) {

	outchan := guild.SystemChannelID
	buffer := fmt.Sprintf("chat $help for commands!")
	_, err := glob.DS.ChannelMessageSend(outchan, buffer)
	if err != nil {
		support.Log("Couldn't send message: stat command")
	}
	return
}

//-------------------------
//-- SHUTDOWN PROCEDURE
//-------------------------

func shutdownchecks() {

	support.Log("Checking if FMDB is dirty...")
	waiting := 0
	for {
		glob.FactModDirtyLock.Lock()

		if glob.FactModDirty == true {
			time.Sleep(100 * time.Millisecond)
			if waiting == 0 {
				support.Log("Waiting for dirty flag to be cleared...")
			} else if waiting%50 == 0 { //about 5 seconds
				support.Log("still waiting...")
			}
			waiting++
		} else {
			glob.FactModDirtyLock.Unlock()
			break
		}

		glob.FactModDirtyLock.Unlock()
	}

	support.Log("Checking if GDB is dirty...")
	waiting = 0
	for {
		glob.GuildDirtyLock.Lock()

		if glob.GuildDirty == true {
			time.Sleep(100 * time.Millisecond)
			if waiting == 0 {
				support.Log("Waiting for dirty flag to be cleared...")
			} else if waiting%50 == 0 { //about 5 seconds
				support.Log("still waiting...")
			}
			waiting++
		} else {
			glob.GuildDirtyLock.Unlock()
			break
		}

		glob.GuildDirtyLock.Unlock()
	}

	support.Log("Waiting for FMDB write lock...")
	glob.FactModWriteLock.Lock()
	support.Log("Waiting for GDB write lock...")
	glob.GuildWriteLock.Lock()
	support.Log("Waiting for expire write lock...")
	glob.ExpireFileLock.Lock()

	support.Log("Waiting for FMDB lock...")
	glob.FactModLock.Lock()
	support.Log("Waiting for GDB lock...")
	glob.GuildLock.Lock()

	support.Log("All locks clear!")
}

//End main loop
