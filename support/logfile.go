package support

import (
	"fmt"
	"os"
	"time"

	"../glob"
)

//Add normal printf to error logs
func CritErrorLog(str string, err error) {
	errorlog, rip := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		panic(rip)
	}
	defer errorlog.Close()
	errorlog.WriteString(fmt.Sprintf("%s %s\n", str, err))
}

func ErrorLog(str string, err error) {
	errorlog, rip := os.OpenFile("error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		//panic(rip)
		return
	}
	defer errorlog.Close()
	errorlog.WriteString(fmt.Sprintf("%s %s\n", str, err))
}

func Log(format string, a ...interface{}) {
	log, rip := os.OpenFile("modwire.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		//panic(rip)
		return
	}
	defer log.Close()

	now := time.Now()
	stamp := now.Format(glob.LogFile)
	stamp = stamp + ": "

	if a != nil {
		log.WriteString(fmt.Sprintf(stamp+format+"%v\n", a...))
		fmt.Println(fmt.Sprintf(stamp+format+" %v", a...))
	} else {
		log.WriteString(stamp + format + "\n")
		fmt.Println(stamp + format)
	}
}

func DebugLog(format string, a ...interface{}) {
	log, rip := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		//panic(rip)
		return
	}
	defer log.Close()

	now := time.Now()
	stamp := now.Format(glob.LogFile)
	stamp = stamp + " DEBUG: "

	if a != nil {
		log.WriteString(fmt.Sprintf(stamp+format+"%v\n", a...))
		fmt.Println(fmt.Sprintf(stamp+format+" %v", a...))
	} else {
		log.WriteString(stamp + format + "\n")
		fmt.Println(stamp + format)
	}
}

func CmdLog(format string, a ...interface{}) {
	log, rip := os.OpenFile("cmd.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// If we encounter an error here, something is seriously wrong.
	if rip != nil {
		//panic(rip)
		return
	}
	defer log.Close()

	now := time.Now()
	stamp := now.Format(glob.LogFile)
	stamp = stamp + " : "

	if a != nil {
		log.WriteString(fmt.Sprintf(stamp+format+"%v\n", a...))
		fmt.Println(fmt.Sprintf(stamp+format+" %v", a...))
	} else {
		log.WriteString(stamp + format + "\n")
		fmt.Println(stamp + format)
	}
}
