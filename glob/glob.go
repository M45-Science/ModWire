package glob

import (
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	//"github.com/sasha-s/go-deadlock"
)

var logger *log.Logger

//-------------------------
//-- GLOBALS --
//-------------------------
const Version = "0.8.11"

//Important globals
var DS *discordgo.Session
var Running bool

//Maximums
const GuildsMax = 32     //After this shars are required
const FactModMax = 16384 //had to icrease this
const ModsMax = 32
const RolesMax = 32
const ChannelsMax = 200
const MaxMembers = 1000
const SearchMax = 25
const AdminsMax = 16

//GLOBAL DATABASES
//Factorio Mod DB
var FactModDB [FactModMax]ModDB
var FactModLock sync.RWMutex
var FactModDBLen = 0

//Guild Database
var GuildDatabase [GuildsMax]GuildFormat
var GuildDatabaseLen = 0
var GuildLock sync.RWMutex

//Dirtys, dirt locks, and write locks
var FactModDirty = false
var FactModDirtyLock sync.RWMutex
var FactModWriteLock sync.Mutex

var GuildDirty = false
var GuildDirtyLock sync.RWMutex
var GuildWriteLock sync.Mutex
var GuildMemberChunkLock sync.Mutex

//Factorio username/tokens
var PlayerName = ""
var Token = ""

//Timestamp of last factorio mod refresh
var LastFactModRefresh time.Time

//Uptime
var BootTime time.Time

//Last dbsave
var LastGDBSave time.Time
var LastFactModSave time.Time
var ExpireFileLock sync.Mutex

//Debug messages
var DEBUG bool = true
var XDEBUG bool = true

//Discord commands
var CL Commands

//-------------------------
//-- SETTINGS --
//-------------------------

//save database in this version format
const FMDBVersion = "0.0.4"
const GDBVersion = "0.0.2"

//-------------------------
//-- STRUCTS --
//-------------------------

//Our mod database
type ModDB struct {
	Name            string
	Title           string
	Thumbnail       string
	Owner           string
	Summary         string
	Downloads_Count int
	Category        string

	Changelog   string
	Description string
	Homepage    string
	UpdatedAt   string
	CreatedAt   string

	Latest Late

	Enabled bool
	Deleted bool
}

//SendModInfo UpdateType
const (
	TypeInfo   = 0
	TypeUpdate = 1
	TypeDelete = 2
	TypeNew    = 3
)

//-------------------------
//-- Factorio Mod Portal JSON
//-------------------------
type FactModJSON struct {
	Results []FMResults
}
type FMResults struct {
	Name            string
	Title           string
	Thumbnail       string
	Owner           string
	Summary         string
	Downloads_Count int
	Category        string
	Latest_release  map[string]interface{}
	Latest          Late

	Deleted bool
}

type Late struct {
	Download_Url string
	File_Name    string
	Info_Json    Ijson
	Released_At  string
	Version      string //db
	Sha1         string
}

type Ijson struct {
	Factorio_Version string //db
}

//-------------------------
//-- Factorio mod portal (detail) JSON
//-------------------------
//Results
type FactModDJSON struct {
	Results []MPMatches
}

//Results contents
type MPMatches struct {
	Name            string
	Title           string
	Thumbnail       string
	ChangeLog       string
	Description     string
	Downloads_Count int
	Category        string
	Score           float32
	Homepage        string
	Updated_At      string
	Created_At      string
	Owner           string
	Releases        []MPReleases

	Deleted bool
}

//MPReleasess
type MPReleases struct {
	Version      string
	Download_Url string
	File_Name    string
}

//-------------------------
//-- Guild Format
//-------------------------
type GuildFormat struct {
	ID          string
	Name        string
	Icon        string
	OwnerID     string
	Splash      string
	MemberCount int
	Large       bool

	Roles       [RolesMax]RoleFormat
	RolesLength int

	Channels       [ChannelsMax]ChanFormat
	ChannelsLength int

	Unavailable bool

	WidgetChannelID string
	SystemChannelID string

	//Local
	PostChannel string
	CmdChannel  string
	BotRole     string

	FactMods    []string
	FactModVers []string

	//Not saved
	Admins    [AdminsMax]GuildAdminFormat
	AdminsLen int

	Deleted bool
}

type GuildAdminFormat struct {
	ID string
}

type ChanFormat struct {
	Name     string
	ID       string
	Position int
	ParentID string

	Deleted bool
}

type RoleFormat struct {
	ID          string
	Name        string
	Managed     bool
	Hoist       bool
	Color       int
	Position    int
	Permissions int64 //bitwise mask

	Deleted bool
}

//-------------------------
//-- player-data-json
//-------------------------
type FPDTokens struct {
	ServiceUsername string `json:"service-username"`
	ServiceToken    string `json:"service-token"`
}

//-------------------------
//-- Discord commands
//-------------------------
// Commands is a struct containing a slice of Command.
type Commands struct {
	CommandList []Command
}

// Command is a struct containing fields that hold command information.
type Command struct {
	Name      string
	Command   func(s *discordgo.Session, m *discordgo.MessageCreate, args []string, channel string)
	Admin     bool
	OwnerOnly bool

	Help  string
	HelpB string
}

//Time formats
const (
	ANSIC       = "Mon Jan _2 15:04:05 2006"
	UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      = "02 Jan 06 15:04 MST"
	RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	RFC3339     = "2006-01-02T15:04:05Z07:00"
	RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     = "3:04PM"
	// Handy time stamps.
	Stamp      = "Jan _2 15:04:05"
	StampMilli = "Jan _2 15:04:05.000"
	StampMicro = "Jan _2 15:04:05.000000"
	StampNano  = "Jan _2 15:04:05.000000000"

	//Custom
	LogFile  = "3:04:05.000--01._2.06"
	FileName = "3-04-05-000-01-02-06"
)
