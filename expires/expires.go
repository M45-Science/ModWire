package expires

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"../glob"
	"../support"
)

func ReadExpireFile() (sleepfor time.Duration) {
	rewrite := false

	//-- LOCK ---
	glob.ExpireFileLock.Lock()
	//-- LOCK ---

	str, err := ioutil.ReadFile(support.Config.DateStampFile)

	//In case of empty file
	if err != nil {
		support.Log("DateStampFile couldn't be read.")
		rewrite = true
	}

	strp := strings.Split(string(str), "\n")

	if len(strp) > 1 {
		t, errb := time.Parse(glob.UnixDate, string(strp[0]))

		if errb != nil {
			t = time.Now()
			support.Log("Unable to parse DateStampFile")
			rewrite = true
		}

		glob.LastFactModRefresh = t
		support.Log("Expire file read.")

	}
	//-- UNLOCK ---
	glob.ExpireFileLock.Unlock()
	//-- UNLOCK ---
	if rewrite {
		WriteExpireFile()
	}
	return
}

func NewsRefreshDelay() (diff time.Duration) {
	t := glob.LastFactModRefresh

	rate, _ := strconv.Atoi(support.Config.FactModRefreshRate)
	ntime := time.Now()

	future := t.Add(time.Duration(rate) * time.Minute)
	diff = future.Sub(ntime)

	if glob.XDEBUG {
		buffer := fmt.Sprintf("rate: %d, t: %s, future: %s, diff: %s", rate, t.String(), future.String(), diff.String())
		support.Log(buffer)
	}

	//Cap incase of clock change
	max := (time.Duration(rate) * time.Minute)

	//Sanity checks
	if diff > max {
		diff = max

		glob.LastFactModRefresh = time.Now()
		support.Log("Mod sleep time over max, capping.")
	} else if diff < 0 {
		diff = (10 * time.Second)
		support.Log("Mod sleep negative, pausing and refreshing.")
	}
	return diff
}

func CacheRefreshDelay() (diff time.Duration) {
	t := glob.LastFactModRefresh

	rate, _ := strconv.Atoi(support.Config.FactModRefreshRate)
	ntime := time.Now()

	future := t.Add(time.Duration(rate) * time.Minute)
	diff = future.Sub(ntime)

	if glob.XDEBUG {
		buffer := fmt.Sprintf("rate: %d, t: %s, future: %s, diff: %s", rate, t.String(), future.String(), diff.String())
		support.Log(buffer)
	}

	//Cap incase of clock change
	max := (time.Duration(rate) * time.Minute)

	//Sanity check
	if diff > max {
		diff = max

		glob.LastFactModRefresh = time.Now()
		support.Log("Search sleep time over max, capping.")
	} else if diff < 0 {
		support.Log("Search sleep negative, pausing and refreshing.")
		diff = (10 * time.Second)
	}
	return diff
}

//End main loop
func WriteExpireFile() {
	//-- LOCK ---
	glob.ExpireFileLock.Lock()
	//-- LOCK ---

	//write last refresh to file, in case we reboot/crash
	ds, err := os.Create(support.Config.DateStampFile)
	if err != nil {
		support.Log("Couldn't create DateStampFile...")
	}
	// close ds on exit and check for its returned error
	defer func() {
		if err := ds.Close(); err != nil {
			panic(err)
		}
	}()

	datestring := fmt.Sprintf("%s\n", glob.LastFactModRefresh.Format(glob.UnixDate))
	errb := ioutil.WriteFile(support.Config.DateStampFile, []byte(datestring), 0644)

	if errb != nil {
		support.Log("Couldn't write DateStampFile..")
	}
	if glob.DEBUG {
		support.Log("Wrote DateStampFile file:", support.Config.DateStampFile)
	}

	//-- UNLOCK ---
	glob.ExpireFileLock.Unlock()
	//-- UNLOCK ---

}
