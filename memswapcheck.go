package main

import (
	"flag"
	"fmt"
	"github.com/olorin/nagiosplugin"
	"github.com/shirou/gopsutil/mem"
	"log"
	"os"
)

func checkerr(e error, action string) {
	if e != nil {
		log.Fatal("Error received while ", action, ": ", e)
	}
}
func debugLogNew(debug bool) func(format string, v ...interface{}) {
	return func(format string, v ...interface{}) {
		if debug {
			log.Printf(format, v...)
		}
	}
}

var debugLog func(format string, v ...interface{})

func main() {
	//Parse CLI options
	var warn float64
	var crit float64
	var d bool
	flag.Float64Var(&warn, "w", 10.0, "Memory used % Warning")
	flag.Float64Var(&crit, "c", 5.0, "Memory used % Critical")
	flag.BoolVar(&d, "d", false, "Enable debug messages")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage of %s: -w <warning> -c <critical> \n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()
	debugLog = debugLogNew(d)
	// Initialize the check - this will return an UNKNOWN result
	// until more results are added.
	check := nagiosplugin.NewCheck()
	// If we exit early or panic() we'll still output a result.
	defer check.Finish()
	// obtain data here
	meminfo, mem_err := mem.VirtualMemory()
	checkerr(mem_err, "getting mem info")
	debugLog("meminfo: %v\n", meminfo)
	swapinfo, swap_err := mem.SwapMemory()
	checkerr(swap_err, "getting swap info")
	debugLog("swapinfo: %v\n", swapinfo)
	totFreePercent := (float64((meminfo.Available + swapinfo.Free)) / float64((meminfo.Total + swapinfo.Total))) * 100
	debugLog("totinfo: %v\n", totFreePercent)
	//Add performance data
	check.AddPerfDatum("MEM USED", "%", meminfo.UsedPercent, 0, 100)
	check.AddPerfDatum("SWAP USED", "%", swapinfo.UsedPercent, 0, 100)
	check.AddPerfDatum("TOT FREE", "%", totFreePercent, 0, 100, warn, crit)
	//Check against thresholds
	switch {
	case warn < crit:
		check.AddResult(nagiosplugin.UNKNOWN, "Warning must be greater than critical")
	case totFreePercent > warn:
		check.AddResult(nagiosplugin.OK, "")
	case totFreePercent < warn && totFreePercent > crit:
		check.AddResult(nagiosplugin.WARNING, "")
	case totFreePercent < crit:
		check.AddResult(nagiosplugin.CRITICAL, "")
	default:
		check.AddResult(nagiosplugin.UNKNOWN, "Unknown state")
	}
}
