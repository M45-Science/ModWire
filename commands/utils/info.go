package utils

import (
	"net/url"
	"strings"

	"../../factorio"
	"../../glob"
	"../../support"
	"github.com/bwmarrin/discordgo"
)

func ModInfo(s *discordgo.Session, m *discordgo.MessageCreate, args []string, channel string) {
	numargs := len(args)

	if numargs < 1 {
		_, err := s.ChannelMessageSend(channel, "Okay, but what mod?")
		if err != nil {
			support.Log("Couldn't send message: modinfo command")
		}
		return
	}

	argstring := strings.Join(args[1:], "%20")
	escapedargs, _ := url.PathUnescape(argstring)
	escapedargs = strings.ReplaceAll(escapedargs, "https://mods.factorio.com/mod/", "")
	cleanurl := &url.URL{Path: escapedargs}
	cleanedargs := cleanurl.String()

	//--- READ-ONLY LOCK ---
	glob.FactModLock.RLock()
	//--- READ-ONLY LOCK ---

	wasfound := false
	iscached := false
	modpos := -1
	for spos := 0; spos < glob.FactModDBLen; spos++ {
		if strings.ToLower(glob.FactModDB[spos].Name) == strings.ToLower(cleanedargs) {
			wasfound = true
			modpos = spos

			if glob.FactModDB[spos].Changelog != "" {
				iscached = true
			} else if glob.FactModDB[spos].Description != "" {
				iscached = true
			}

			if iscached {
				support.SendModInfo(glob.FactModDB[spos], glob.TypeInfo, channel)
				//--- READ-ONLY UNLOCK ---
				glob.FactModLock.RUnlock()
				//--- READ-ONLY UNLOCK ---
				return
			}
		}
	}

	//Found it, but don't have details...
	if wasfound && modpos >= 0 {
		//-- RUNLOCK ---
		glob.FactModLock.RUnlock()
		glob.FactModLock.Lock()
		//-- LOCK

		//Expects RW locked moddb unlocks when done.
		factorio.GetDetails(glob.FactModDB[modpos].Name, modpos)
		//-- UNLOCK ---
		glob.FactModLock.Unlock()
		glob.FactModLock.RLock()
		//--- READ-ONLY LOCK ---
		support.SendModInfo(glob.FactModDB[modpos], glob.TypeInfo, channel)
	} else {
		_, err := s.ChannelMessageSend(channel, "Couldn't find a mod called that.")
		if err != nil {
			support.Log("Couldn't send message: info command")
		}
		//-- RUNLOCK ---
		glob.FactModLock.RUnlock()
		_, errb := s.ChannelMessageSend(channel, factorio.ModSearch(cleanedargs))
		if errb != nil {
			support.Log("Couldn't send message: stat command")
		}
		glob.FactModLock.RLock()
		//--- READ-ONLY LOCK ---
	}

	//--- READ-ONLY UNLOCK ---
	glob.FactModLock.RUnlock()
	//--- READ-ONLY UNLOCK ---
	return
}
