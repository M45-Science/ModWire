package support

import (
	"log"
	"os"
	"strings"

	"../glob"
	"github.com/joho/godotenv"
)

// Config is a config interface.
var Config config

type config struct {
	//General
	FMDBFile         string
	GDBFile          string
	Verbose          string
	DateStampFile    string
	MaxFetchAttempts string
	RetryMultiplier  string
	TwitchURL        string
	RoleName         string

	//Discord-Specific
	DiscordToken string

	//Legacy
	DiscordCmdChannel  string
	DiscordPostChannel string
	DiscordAllChannel  string
	DiscordAdminIDs    []string

	DiscordCommandPrefix string
	//Rename These
	DiscordMaxLineLength string
	DiscordMaxLines      string
	DiscordMaxTitleWords string

	//Factorio-Specific
	FactModRefreshRate string
	//FactModDLFolder        string
	//FactMaxModDLTries      string
	FactPlayerDataLocation string

	FactModUrl     string
	FactModDLUrl   string
	FactVersionUrl string
}

func (conf *config) LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Config = config{
		//General
		FMDBFile:         os.Getenv("FMDBFile"),
		GDBFile:          os.Getenv("GDBFile"),
		Verbose:          os.Getenv("Verbose"),
		DateStampFile:    os.Getenv("DateStampFile"),
		MaxFetchAttempts: os.Getenv("MaxFetchAttempts"),
		RetryMultiplier:  os.Getenv("RetryMultiplier"),
		TwitchURL:        os.Getenv("TwitchURL"),
		RoleName:         os.Getenv("RoleName"),

		//Discord-Specific
		DiscordToken: os.Getenv("DiscordToken"),

		//Legacy
		DiscordCmdChannel:  os.Getenv("DiscordCmdChannel"),
		DiscordPostChannel: os.Getenv("DiscordPostChannel"),
		DiscordAllChannel:  os.Getenv("DiscordAllChannel"),
		DiscordAdminIDs:    strings.Split(os.Getenv("DiscordAdminIDs"), " "),

		DiscordCommandPrefix: os.Getenv("DiscordCommandPrefix"),
		//Rename These
		DiscordMaxLineLength: os.Getenv("DiscordMaxLineLength"),
		DiscordMaxLines:      os.Getenv("DiscordMaxLines"),
		DiscordMaxTitleWords: os.Getenv("DiscordMaxTitleWords"),

		//Factorio-Specific
		FactModRefreshRate: os.Getenv("FactModRefreshRate"),
		//FactModDLFolder:        os.Getenv("FactModDLFolder"),
		//FactMaxModDLTries:      os.Getenv("FactMaxModDLTries"),
		FactPlayerDataLocation: os.Getenv("FactPlayerDataLocation"),
		FactModUrl:             os.Getenv("FactModUrl"),
		FactModDLUrl:           os.Getenv("FactModDLUrl"),
		FactVersionUrl:         os.Getenv("FactVersionUrl"),
	}
	if strings.ToLower(Config.Verbose) == "true" {
		glob.DEBUG = true
	} else {
		glob.DEBUG = false
	}
	Log("Envioment loaded.")
}
