package database

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	b64 "encoding/base64"

	"../expires"
	"../glob"
	"../support"
)

//-------------------------
//-- DATABASE --
//-------------------------

//Save Search DB
func SaveFMDB() {
	//-- LOCK ---
	glob.FactModWriteLock.Lock()
	defer glob.FactModWriteLock.Unlock()
	//-- DEFER ---

	buffer := ""
	oldname := support.Config.FMDBFile
	newname := fmt.Sprintf("dbbackups/FMDB-%s.bak", time.Now().Format(glob.FileName))

	err := os.Rename(oldname, newname)
	if err != nil {
		support.Log("Couldn't backup FMDB.")
	}

	// open output file
	fo, err := os.Create(support.Config.FMDBFile)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	if glob.FMDBVersion == "0.0.4" {
		buffer = buffer + fmt.Sprintf("FMDB,0.0.4:")
		//--- RLOCK ---
		glob.FactModLock.RLock()
		//--- RLOCK ---
		for pos := 0; pos < glob.FactModDBLen; pos++ {
			if pos < glob.FactModMax {
				tmp := fmt.Sprintf("%d", glob.FactModDB[pos].Downloads_Count)
				buffer = buffer + fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s:",
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Name)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Title)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Thumbnail)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Owner)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Summary)),
					b64.StdEncoding.EncodeToString([]byte(tmp)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Category)),

					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Changelog)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Description)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Homepage)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].UpdatedAt)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].CreatedAt)),

					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Latest.Download_Url)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Latest.File_Name)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Latest.Info_Json.Factorio_Version)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Latest.Released_At)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Latest.Version)),
					b64.StdEncoding.EncodeToString([]byte(glob.FactModDB[pos].Latest.Sha1)))
			}

		}
		//--- RUNLOCK ---
		glob.FactModLock.RUnlock()
		//--- RUNLOCK ---

		err = ioutil.WriteFile(support.Config.FMDBFile, []byte(buffer), 0644)

		if err != nil {
			support.ErrorLog("FMDB File write error: ", err)
		}

		support.Log("FMDB wrote to ", support.Config.FMDBFile)
		glob.LastFactModRefresh = time.Now()
		expires.WriteExpireFile()
		return
	} else {
		support.Log("FMDB format invalid, critical error.")
		os.Exit(1)
		return
	}
}

//Load Search DB
func LoadFMDB() {
	//-- LOCK ---
	glob.FactModWriteLock.Lock()
	defer glob.FactModWriteLock.Unlock()
	//-- DEFER ---

	filedata, err := ioutil.ReadFile(support.Config.FMDBFile)
	if err != nil {
		support.Log("Error reading FMDB.")
		return
	}

	if filedata != nil {

		glob.FactModDBLen = 0
		dblines := strings.Split(string(filedata), ":")
		numlines := len(dblines) - 1
		dbversion := ""

		//--- LOCK ---
		glob.FactModLock.Lock()
		defer glob.FactModLock.Unlock()
		//--- DEFER ---

		for pos := 0; pos < numlines; pos++ {
			if pos < glob.FactModMax {
				items := strings.Split(string(dblines[pos]), ",")
				if pos == 0 {
					if items[0] != "FMDB" {
						support.Log("invalid FMDB file.")
						return
					} else {
						dbversion = items[1]
						if glob.XDEBUG {
							support.Log("FMDB version: ", dbversion)
						}
					}
				} else {
					if dbversion == "0.0.3" {
						tmp, _ := b64.StdEncoding.DecodeString(items[0])
						glob.FactModDB[glob.FactModDBLen].Name = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[1])
						glob.FactModDB[glob.FactModDBLen].Title = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[2])
						tmpint, _ := strconv.ParseInt(string(tmp), 10, 16)
						glob.FactModDB[glob.FactModDBLen].Downloads_Count = int(tmpint)

						tmp, _ = b64.StdEncoding.DecodeString(items[3])
						glob.FactModDB[glob.FactModDBLen].Latest.Info_Json.Factorio_Version = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[4])
						glob.FactModDB[glob.FactModDBLen].Latest.Version = string(tmp)

						if glob.XDEBUG {
							support.Log("(FMDB) loading ", glob.FactModDB[glob.FactModDBLen].Name)
						}

						glob.FactModDBLen++
					} else if dbversion == "0.0.4" {
						tmp, _ := b64.StdEncoding.DecodeString(items[0])
						glob.FactModDB[glob.FactModDBLen].Name = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[1])
						glob.FactModDB[glob.FactModDBLen].Title = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[2])
						glob.FactModDB[glob.FactModDBLen].Thumbnail = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[3])
						glob.FactModDB[glob.FactModDBLen].Owner = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[4])
						glob.FactModDB[glob.FactModDBLen].Summary = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[5])
						tmpint, _ := strconv.ParseInt(string(tmp), 10, 32)
						glob.FactModDB[glob.FactModDBLen].Downloads_Count = int(tmpint)

						tmp, _ = b64.StdEncoding.DecodeString(items[6])
						glob.FactModDB[glob.FactModDBLen].Category = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[7])
						glob.FactModDB[glob.FactModDBLen].Changelog = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[8])
						glob.FactModDB[glob.FactModDBLen].Description = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[9])
						glob.FactModDB[glob.FactModDBLen].Homepage = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[10])
						glob.FactModDB[glob.FactModDBLen].UpdatedAt = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[11])
						glob.FactModDB[glob.FactModDBLen].CreatedAt = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[12])
						glob.FactModDB[glob.FactModDBLen].Latest.Download_Url = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[13])
						glob.FactModDB[glob.FactModDBLen].Latest.File_Name = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[14])
						glob.FactModDB[glob.FactModDBLen].Latest.Info_Json.Factorio_Version = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[15])
						glob.FactModDB[glob.FactModDBLen].Latest.Released_At = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[16])
						glob.FactModDB[glob.FactModDBLen].Latest.Version = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[17])
						glob.FactModDB[glob.FactModDBLen].Latest.Sha1 = string(tmp)

						if glob.XDEBUG {
							//support.Log("(FMDB) loading ", glob.FactModDB[glob.FactModDBLen].Name)
						}

						glob.FactModDBLen++
					} else {
						support.Log("Incompatable FMDB version.")
						return
					}
				}
			} else {
				support.Log("Couldn't load all FMDB, limit reached.")
				break
			}
		}
		//If DB isn't current version... resave it in new format after we check web.
		if dbversion != glob.FMDBVersion {
			//Mark as dirty
			glob.FactModDirtyLock.Lock()
			glob.FactModDirty = true
			glob.FactModDirtyLock.Unlock()
		}
	}
	support.Log("FMDB read from", support.Config.FMDBFile)
}

//---------------
// Guild DB
//--------------
func SaveGuilds() {
	//-- LOCK ---
	glob.GuildWriteLock.Lock()
	defer glob.GuildWriteLock.Unlock()
	//-- DEFER ---

	buffer := ""
	oldname := support.Config.GDBFile
	newname := fmt.Sprintf("dbbackups/GDB-%s.bak", time.Now().Format(glob.FileName))

	err := os.Rename(oldname, newname)
	if err != nil {
		support.Log("Couldn't backup GDB.")
	}

	// open output file
	fo, err := os.Create(support.Config.GDBFile)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	if glob.GDBVersion == "0.0.2" {
		buffer = buffer + fmt.Sprintf("GDB,0.0.2:")
		//--- RLOCK ---
		glob.GuildLock.RLock()
		//--- RLOCK ---
		for pos := 0; pos < glob.GuildDatabaseLen; pos++ {
			if pos < glob.GuildsMax {
				tmp := ""
				if glob.GuildDatabase[pos].Deleted {
					tmp = "true"
				} else {
					tmp = "false"
				}

				buffer = buffer + fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s:",
					b64.StdEncoding.EncodeToString([]byte(glob.GuildDatabase[pos].Name)),
					b64.StdEncoding.EncodeToString([]byte(glob.GuildDatabase[pos].ID)),
					b64.StdEncoding.EncodeToString([]byte(glob.GuildDatabase[pos].PostChannel)),
					b64.StdEncoding.EncodeToString([]byte(glob.GuildDatabase[pos].CmdChannel)),
					b64.StdEncoding.EncodeToString([]byte(glob.GuildDatabase[pos].BotRole)),
					b64.StdEncoding.EncodeToString([]byte(strings.Join(glob.GuildDatabase[pos].FactMods, ","))),
					b64.StdEncoding.EncodeToString([]byte(strings.Join(glob.GuildDatabase[pos].FactModVers, ","))),
					b64.StdEncoding.EncodeToString([]byte(tmp)))
			}

		}
		//--- RUNLOCK ---
		glob.GuildLock.RUnlock()
		//--- RUNLOCK ---

		err = ioutil.WriteFile(support.Config.GDBFile, []byte(buffer), 0644)

		if err != nil {
			support.ErrorLog("GDB File write error: ", err)
		}

		support.Log("GDB wrote to ", support.Config.GDBFile)
		return
	} else {
		support.Log("GDB format invalid, critical error.")
		os.Exit(1)
		return
	}
}

func LoadGuilds() {
	//-- LOCK ---
	glob.GuildWriteLock.Lock()
	defer glob.GuildWriteLock.Unlock()
	//-- DEFER ---

	filedata, err := ioutil.ReadFile(support.Config.GDBFile)
	if err != nil {
		support.Log("Error reading GDB.")
		return
	}

	if filedata != nil {

		glob.GuildDatabaseLen = 0
		dblines := strings.Split(string(filedata), ":")
		numlines := len(dblines) - 1
		dbversion := ""

		//--- LOCK ---
		glob.GuildLock.Lock()
		defer glob.GuildLock.Unlock()
		//--- DEFER ---

		for pos := 0; pos < numlines; pos++ {
			if pos < glob.GuildsMax {
				items := strings.Split(string(dblines[pos]), ",")
				if pos == 0 {
					if items[0] != "GDB" {
						support.Log("invalid GDB file.")
						return
					} else {
						dbversion = items[1]
						if glob.XDEBUG {
							support.Log("GDB version: ", dbversion)
						}
					}
				} else {
					if dbversion == "0.0.2" {
						tmp, _ := b64.StdEncoding.DecodeString(items[0])
						glob.GuildDatabase[glob.GuildDatabaseLen].Name = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[1])
						glob.GuildDatabase[glob.GuildDatabaseLen].ID = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[2])
						glob.GuildDatabase[glob.GuildDatabaseLen].PostChannel = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[3])
						glob.GuildDatabase[glob.GuildDatabaseLen].CmdChannel = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[4])
						glob.GuildDatabase[glob.GuildDatabaseLen].BotRole = string(tmp)

						tmp, _ = b64.StdEncoding.DecodeString(items[5])
						tstrings := strings.Split(string(tmp), ",")
						glob.GuildDatabase[glob.GuildDatabaseLen].FactMods = tstrings

						tmp, _ = b64.StdEncoding.DecodeString(items[6])
						tstrings = strings.Split(string(tmp), ",")
						glob.GuildDatabase[glob.GuildDatabaseLen].FactModVers = tstrings

						tmp, _ = b64.StdEncoding.DecodeString(items[7])
						if string(tmp) == "true" {
							glob.GuildDatabase[glob.GuildDatabaseLen].Deleted = true
						} else {
							glob.GuildDatabase[glob.GuildDatabaseLen].Deleted = false
						}

						if glob.XDEBUG {
							support.Log("(GDB) loading ", glob.GuildDatabase[glob.GuildDatabaseLen].Name)
						}

						glob.GuildDatabaseLen++
					} else {
						support.Log("Incompatable GDB version.")
						return
					}
				}
			} else {
				support.Log("Couldn't load all GDB, limit reached.")
				break
			}
		}
		//If DB isn't current version... resave it in new format after we check web.
		if dbversion != glob.GDBVersion {
			//Mark as dirty
			glob.GuildDirtyLock.Lock()
			glob.GuildDirty = true
			glob.GuildDirtyLock.Unlock()
		}
	}
	support.Log("GDB read from", support.Config.GDBFile)
}
