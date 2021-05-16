package main

import (
	"fmt"
	"io"
	"os/exec"
)

//------------------------------------------------------------------------------

func notPass(message string) {
	fmt.Printf("%sDO NOT PASS%s: %s\n", colorRed, colorDefault, message)
}

func runCmdWithStdin(stdinStr string, cmdName string) (cmdOutput string, exitCode int) {
	//fmt.Printf("%s %s\n", stdinStr, cmdName)
	var cmd = exec.Command(cmdName)
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
