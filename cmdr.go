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

var wg sync.WaitGroup

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

	for i := 0; i < processes; i++ {
		wg.Add(1)
		go runCommand(i, command, staggered, restart)
	}

	wg.Wait()

	fmt.Println("Complete.")
}

func runCommand(id int, command string, staggered bool, restart bool) {
	if staggered {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	}

	split := strings.SplitN(command, " ", 2)

	process := exec.Command(split[0], split[1])

	fmt.Println(
		fmt.Sprintf(
			"\033[36m ==>\033[37m Starting process\033[32m [%d]\033[32m [%s]\033[37m",
			id,
			command,
		),
	)

	process.Stdout = os.Stdout
	process.Stderr = os.Stderr

	if err := process.Run(); err != nil {
		defer wg.Done()

		fmt.Println(fmt.Sprintf("\033[31m ==> Error: [%d] %s \033[37m", id, err.Error()))
	} else if restart {
		fmt.Printf("\033[31m ==> Process [%d] stopped, possibly due to queue restart signal. Restarting...\033[37m\n", id)

		runCommand(id, command, staggered, restart)
	} else {
		defer wg.Done()
	}
}
