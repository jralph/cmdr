package main

import (
	"fmt"
	"github.com/pborman/getopt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	processes := getopt.IntLong("processes", 'n', 1, "The number of processes to run.")
	staggered := getopt.BoolLong("staggered", 's', "", "Stagger the starting of processes.")
	restart := getopt.BoolLong("restart", 'r', "", "Restart the command on exit (not failures).")

	opts := getopt.CommandLine

	opts.Parse(os.Args)

	command := opts.Arg(0)

	runCommands(command, *processes, *staggered, *restart)
}

func runCommands(command string, processes int, staggered bool, restart bool) {
	fmt.Println(" ==> Starting processes ...")

	var wg sync.WaitGroup

	for i := 0; i < processes; i++ {
		wg.Add(1)

		var cmd *exec.Cmd

		if strings.Contains(command, " ") {
			commandParts := strings.SplitN(command, " ", 2)

			cmd = exec.Command(commandParts[0], commandParts[1])
		} else {
			cmd = exec.Command(command)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		go runCommand(i, cmd, staggered, restart, &wg)
	}

	wg.Wait()

	fmt.Println("Complete.")
}

func runCommand(id int, command *exec.Cmd, staggered bool, restart bool, wg *sync.WaitGroup) {
	if staggered {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}

	fmt.Println(
		fmt.Sprintf(
			"\033[36m ==>\033[37m Starting process\033[32m [%d]\033[32m [%s]\033[37m",
			id,
			strings.Join(command.Args, " "),
		),
	)

	if err := command.Run(); err != nil {
		defer wg.Done()

		fmt.Println(fmt.Sprintf("\033[31m ==> Error: [%d] %s \033[37m", id, err.Error()))
	} else if restart {
		fmt.Printf("\033[31m ==> Process [%d] stopped, possibly due to queue restart signal. Restarting...\033[37m\n", id)

		runCommand(id, command, staggered, restart, wg)
	} else {
		defer wg.Done()
	}
}
