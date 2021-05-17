package util

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

//------------------------------------------------------------------------------

var exitCode = 0

func GetExitCode() int {
	return exitCode
}

//------------------------------------------------------------------------------

var colorInfo = "\033[37m"    //light gray
var colorNotPass = "\033[31m" //Red
var colorWarn = "\033[33m"    //Orange/yellow
var colorDebug = "\033[36m"   // Cyan
var colorDefault = "\033[39m"

func DisableColor() {
	colorInfo = ""
	colorNotPass = ""
	colorWarn = ""
	colorDebug = ""
	colorDefault = ""
}

//------------------------------------------------------------------------------

// WildcardSet is an object which stores several strings, each can have a wildcard.
type WildcardSet map[string]int

// NewWildcardSet makes a WildcardSet for using with InWildcardSet().
// Each input values shall have zero or one wildcard character '*'.
func NewWildcardSet(input []string) WildcardSet {
	var ws = make(map[string]int)
	for _, element := range input {
		ws[element] = strings.Index(element, "*")
	}
	return ws
}

// InWildcardSet checks if str matches an element in the WildcardSet.
func InWildcardSet(str string, ws WildcardSet) bool {
	if ws != nil {
		if wildcardIndex, ok := ws[str]; ok {
			if wildcardIndex == -1 { // If the value in the map is -1, it means that it has no wildcard.
				return true
			}
		}
		for key, wildcardIndex := range ws {
			// If the value in the map is !=-1, it means that the key has one character '*'
			// which is a wilcard when finding str.
			if wildcardIndex != -1 {
				var beginning = key[:wildcardIndex]
				var end = key[wildcardIndex+1:]
				if len(str) >= len(beginning)+len(end) &&
					(len(beginning) == 0 || beginning == str[:len(beginning)]) &&
					(len(end) == 0 || end == str[len(str)-len(end):]) {
					//fmt.Printf("matches: %s %s\n", str, key)
					return true
				}
			}
		}
	}
	return false
}

//------------------------------------------------------------------------------

func printWithPrefix(color string, prefix string, message string, args ...interface{}) {
	var str = fmt.Sprintf(message, args...)
	var i = 0
	for _, line := range strings.Split(str, "\n") {
		if i > 0 {
			line = "     |_ " + line
		}
		fmt.Printf(color + prefix + colorDefault + ": " + line + "\n")
		i++
	}
}

func NotPass(message string, args ...interface{}) {
	printWithPrefix(colorNotPass, "INVALID", message, args...)
	exitCode = 2
}

func Info(message string, args ...interface{}) {
	printWithPrefix(colorInfo, "INFO   ", message, args...)
}

func Warn(message string, args ...interface{}) {
	printWithPrefix(colorWarn, "WARNING", message, args...)
}

func DebugPrintf(message string, args ...interface{}) {
	printWithPrefix(colorDebug, "DEBUG  ", message, args...)
}

//------------------------------------------------------------------------------

func RunCmdWithStdin(stdinStr string, cmdName string, cmdArgs []string) (cmdOutput string, exitCode int) {
	//fmt.Printf("%s %s\n", stdinStr, cmdName)
	var cmd = exec.Command(cmdName, cmdArgs...)
	var stdin, err = cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, stdinStr)
	}()

	var out []byte
	out, err = cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			panic(err)
		}
	}

	cmdOutput = string(out)
	return
}

//------------------------------------------------------------------------------
