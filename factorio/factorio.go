package factorio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"../expires"
	"../glob"
	"../support"
	"github.com/mitchellh/mapstructure"
)

//We do this, so we don't break everything if JSON changes, or we add more games
func FMResultsToModDB(in glob.MPMatches) (out glob.ModDB) {

	out.Name = in.Name
	out.Title = in.Title
	out.Thumbnail = in.Thumbnail
	out.Changelog = in.ChangeLog
	out.Description = in.Description
	out.Downloads_Count = in.Downloads_Count
	out.Category = in.Category
	out.Homepage = in.Homepage
	out.UpdatedAt = in.Updated_At
	out.CreatedAt = in.Created_At
	out.Owner = in.Owner

	rel := len(in.Releases) - 1
	out.Latest.Download_Url = in.Releases[rel].Download_Url
	out.Latest.Version = in.Releases[rel].Version
	out.Latest.File_Name = in.Releases[rel].File_Name

	out.Changelog = in.ChangeLog

	return out
}

func ModSearch(arg string) (rstring string) {
	var results [glob.SearchMax]string
	rpos := 0

	support.Log("Searching database for: ", arg)

	//--- READ-ONLY LOCK ---
	glob.FactModLock.RLock()
	//--- READ-ONLY LOCK ---

	//Deeper search
	length := glob.FactModDBLen
	for pos := 0; pos < length; pos++ {
		if rpos < glob.SearchMax {
			if (strings.Contains(strings.ToLower(glob.FactModDB[pos].Name), strings.ToLower(arg)) && glob.FactModDB[pos].Name != "") ||
				(strings.Contains(strings.ToLower(glob.FactModDB[pos].Title), strings.ToLower(arg)) && glob.FactModDB[pos].Title != "") {
				results[rpos] = fmt.Sprintf("%-7d: %s", glob.FactModDB[pos].Downloads_Count, glob.FactModDB[pos].Name)
				rpos++
			}
		} else {
			break
		}
	}

	//--- READ-ONLY UNLOCK ---
	glob.FactModLock.RUnlock()
	//--- READ-ONLY UNLOCK ---

	if rpos > 0 {
		output := ""
		for x := 0; x < rpos; x++ {
			if results[x] != "" {
				output = output + fmt.Sprintf("%s\n", results[x])
			}
		}
		rstring = fmt.Sprintf("Possible Matches: (Sorted by downloads)```\n%s\n```", output)
		if glob.XDEBUG {
			support.Log(rstring)
		}
		return rstring
	} else {
		return "No matching results."
	}
}

//--------
//Get username/token
//--------
func Get_Username_Token() {

	filedata, err := ioutil.ReadFile(support.Config.FactPlayerDataLocation)
	if err == nil {
		if filedata == nil {
			support.Log("Empty file: ", support.Config.FactPlayerDataLocation)
		} else {
			if glob.XDEBUG {
				support.Log("Read file: ", support.Config.FactPlayerDataLocation)
			}
			res := &glob.FPDTokens{}
			err = json.Unmarshal([]byte(filedata), res)
			glob.PlayerName = res.ServiceUsername
			glob.Token = res.ServiceToken
			if err != nil {
				support.Log("Couldn't unmarshal player name/token data.")
				return
			}

			if glob.XDEBUG {
				buffer := fmt.Sprintf("user/token: %s/%s", glob.PlayerName, glob.Token)
				support.Log(buffer)
			}
			if glob.DEBUG {
				support.Log("Player token read.")
			}
		}
	} else {
		support.Log(fmt.Sprintf("Unable to read file:", support.Config.FactPlayerDataLocation))
	}
}

//--- INCOMPLETE ---
func Fetch_Factorio_Versions() {
	versionurl := fmt.Sprintf("https://updater.factorio.com/get-available-versions?username=%s&token=%s&api_version=2",
		url.QueryEscape(glob.PlayerName), url.QueryEscape(glob.Token), "")

	if glob.XDEBUG {
		support.Log("Requesting ", versionurl)
	}

	resp, err := http.Get(versionurl)
	if err != nil {
		support.Log("Couldn't get factorio versions from:", versionurl, "... Retrying.")
		time.Sleep(10 * time.Second)
		Fetch_Factorio_Versions()
		return
	}
	defer resp.Body.Close()
}

type byDownload []glob.ModDB

func (a byDownload) Len() int {
	return len(a)
}
func (a byDownload) Less(i, j int) bool {
	return a[i].Downloads_Count > a[j].Downloads_Count
}
func (a byDownload) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func FetchFactMods(silent bool) (success bool) {

	modurl := "https://mods.factorio.com/api/mods?page_size=1000000"
	if glob.DEBUG {
		support.Log("Requesting full mod portal listing", modurl)
	}

	multi, _ := strconv.Atoi(support.Config.RetryMultiplier)
	maxatt, _ := strconv.Atoi(support.Config.MaxFetchAttempts)
	var resp *http.Response
	var err error

	for att := 1; att < maxatt; att++ {

		resp, err = http.Get(modurl)
		if err != nil {
			support.Log("Couldn't get URL. Attempting again in %d seconds. url/sleep ", modurl, (att * multi))
			time.Sleep(time.Duration(multi*att) * time.Second)
		} else {
			defer resp.Body.Close()
			break
		}

	}

	if resp.Body == nil {
		support.Log("Max fetch attempts, or empty response for:", modurl)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	var modportalresponse string = string(body)

	if err != nil {
		support.ErrorLog("", err)
		return false
	} else {
		res := &glob.FactModJSON{}
		err := json.Unmarshal([]byte(modportalresponse), res)
		if err != nil {
			support.Log("json unmarshal error: ", err)
			panic(err)
			return false
		}

		if len(res.Results) == 0 {
			support.Log("No results in json data.")
			return false
		}

		//-- LOCK ---
		glob.FactModLock.Lock()
		//-- LOCK ---
		dblenth := glob.FactModDBLen

		found := false
		for _, r := range res.Results {
			found = false
			for pos := 0; pos < dblenth; pos++ {
				if glob.FactModDB[pos].Name == r.Name {
					found = true
					glob.FactModDB[pos].Name = r.Name
					glob.FactModDB[pos].Title = r.Title
					glob.FactModDB[pos].Owner = r.Owner
					glob.FactModDB[pos].Summary = r.Summary
					if r.Downloads_Count > glob.FactModDB[pos].Downloads_Count {
						glob.FactModDB[pos].Downloads_Count = r.Downloads_Count
					}
					glob.FactModDB[pos].Category = r.Category
					//Decode into struct, dont know why unmarshal can't do this on its own...
					oldvers := glob.FactModDB[pos].Latest.Version
					input := r.Latest_release
					_ = mapstructure.Decode(input, &glob.FactModDB[pos].Latest)
					newvers := glob.FactModDB[pos].Latest.Version
					glob.FactModDB[pos].Thumbnail = r.Thumbnail

					if glob.FactModDB[pos].Deleted {
						glob.FactModDB[pos].Deleted = false
						glob.FactModDirtyLock.Lock()
						glob.FactModDirty = true
						glob.FactModDirtyLock.Unlock()
						if !silent {
							support.Log("Mod undeleted: ", glob.FactModDB[pos].Name)
							GetDetails((glob.FactModDB[pos].Name), pos)
							support.SendModInfo(glob.FactModDB[pos], glob.TypeNew, support.Config.DiscordAllChannel)
						}
					} else {
						if oldvers != newvers {
							glob.FactModDirtyLock.Lock()
							glob.FactModDirty = true
							glob.FactModDirtyLock.Unlock()
							if !silent {
								support.Log("Mod updated: ", glob.FactModDB[pos].Name)
								GetDetails(glob.FactModDB[pos].Name, pos)
								support.SendModInfo(glob.FactModDB[pos], glob.TypeUpdate, support.Config.DiscordAllChannel)
							}
						}
					}
				}
			}
			if !found {
				//New mod
				glob.FactModDB[glob.FactModDBLen].Name = r.Name
				glob.FactModDB[glob.FactModDBLen].Title = r.Title
				glob.FactModDB[glob.FactModDBLen].Owner = r.Owner
				glob.FactModDB[glob.FactModDBLen].Summary = r.Summary
				glob.FactModDB[glob.FactModDBLen].Downloads_Count = r.Downloads_Count
				glob.FactModDB[glob.FactModDBLen].Category = r.Category
				//Decode into struct, dont know why unmarshal can't do this on its own...
				input := r.Latest_release
				_ = mapstructure.Decode(input, &glob.FactModDB[glob.FactModDBLen].Latest)

				glob.FactModDB[glob.FactModDBLen].Thumbnail = r.Thumbnail
				glob.FactModDB[glob.FactModDBLen].Deleted = false
				glob.FactModDBLen++
				glob.FactModDirtyLock.Lock()
				glob.FactModDirty = true
				glob.FactModDirtyLock.Unlock()
				if !silent {

					support.Log("New mod: ", r.Name)
					//This will eventually intergrate announcing to all servers subscribed
					GetDetails(glob.FactModDB[glob.FactModDBLen-1].Name, glob.FactModDBLen-1)
					support.SendModInfo(glob.FactModDB[glob.FactModDBLen-1], glob.TypeNew, support.Config.DiscordAllChannel)
				}
			}
		}
		found = false

		dblenth = glob.FactModDBLen
		for pos := 0; pos < dblenth; pos++ {
			for _, r := range res.Results {
				if glob.FactModDB[pos].Name == r.Name {
					found = true
				}
			}
			if !found {
				//Deleted mod
				glob.FactModDB[pos].Deleted = true

				//Mark as dirty
				glob.FactModDirtyLock.Lock()
				glob.FactModDirty = true
				glob.FactModDirtyLock.Unlock()

				if !silent {
					support.Log("Mod deleted: ", glob.FactModDB[pos].Name)
					GetDetails(glob.FactModDB[pos].Name, pos)
					support.SendModInfo(glob.FactModDB[pos], glob.TypeDelete, support.Config.DiscordAllChannel)

				}
			}
		}

		sort.Sort(byDownload(glob.FactModDB[:]))
		//-- UNLOCK ---
		glob.FactModLock.Unlock()
		//-- UNLOCK ---
	}
	support.Log("mod database update complete.")
	expires.WriteExpireFile()

	return true
}

func GetDetails(name string, pos int) {

	//Expects locked database
	//-- UNLOCK ---
	glob.FactModLock.Unlock()
	defer glob.FactModLock.Lock()
	//-- DEFER ---

	modurl := fmt.Sprintf("https://mods.factorio.com/api/mods?page_size=max&full=True&namelist=%s", url.QueryEscape(name))
	if glob.XDEBUG {
		support.Log("Requesting ", modurl)
	}

	multi, _ := strconv.Atoi(support.Config.RetryMultiplier)
	maxatt, _ := strconv.Atoi(support.Config.MaxFetchAttempts)

	var resp *http.Response
	var err error

	for att := 1; att < maxatt; att++ {
		resp, err = http.Get(modurl)
		if err != nil {
			support.Log("Couldn't get: attempting again: url/sleep ", modurl, att*multi)
			time.Sleep(time.Duration(multi*att) * time.Second)
		}

		if resp.Body == nil {
			support.Log("Max fetch attempts, or empty response for:", modurl)

			return
		} else {
			defer resp.Body.Close()
			break
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	var modportalresponse string = string(body)

	if err != nil {
		support.ErrorLog("", err)
	} else {
		res := &glob.FactModDJSON{}
		err := json.Unmarshal([]byte(modportalresponse), res)
		if err != nil {
			support.Log("json unmarshal error: ", err)
			return
		}

		if len(res.Results) == 0 {
			support.Log("No results in json data.")
			return
		}

		if name == res.Results[0].Name {
			//-- LOCK ---
			glob.FactModLock.Lock()
			//-- LOCK ---
			for pos := 0; pos < glob.FactModDBLen; pos++ {
				if glob.FactModDB[pos].Name == name {
					glob.FactModDB[pos].Changelog = res.Results[0].ChangeLog
					glob.FactModDB[pos].Description = res.Results[0].Description
					glob.FactModDB[pos].Homepage = res.Results[0].Homepage
					support.Log("Details added for: ", name)

					//Mark dirty
					glob.FactModDirtyLock.Lock()
					glob.FactModDirty = true
					glob.FactModDirtyLock.Unlock()
					break
				}
			}
			//-- UNLOCK ---
			glob.FactModLock.Unlock()
			//-- UNLOCK ---
		} else {
			support.Log("Mod details info had wrong mod name...?")
		}
	}
	return
}
