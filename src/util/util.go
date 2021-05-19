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

var printInfo bool
var printDebug bool
var printWarn bool

func SetVerbosity(printInfoP bool, printDebugP bool, printWarnP bool) {
	printInfo = printInfoP
	printDebug = printDebugP
	printWarn = printWarnP
}

func IsDebug() bool {
	return printDebug
}

func IsInfo() bool {
	return printInfo
}

func IsWarn() bool {
	return printWarn
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
