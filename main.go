package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Process struct {
	PID   int
	Name  string
	State string
	PPID  int
}

func main() {

	args := os.Args
	if len(args) > 1 {
		if args[1] == "inspect" {
			pid := os.Getpid()
			fmt.Println(pid)

			ppid := os.Getppid() // Parent PID
			fmt.Println("PPID:", ppid)

			cwd, _ := os.Getwd() // Current working directory
			fmt.Println("CWD:", cwd)

		}
	}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("TinyTop started")
	fmt.Println("Commands: list, stop <pid>, continue <pid>, terminate <pid>, quit")

	for {
		fmt.Print("tinytop> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		fmt.Println(input)
	}

	ps, _ := os.ReadDir("/proc")

	var processes []Process

	for p := range ps {
		MyProcess, err := ProcessParser(p)
		if err != nil {
			continue
		} else {
			processes = append(processes, MyProcess)
		}

	}

	PrintProcess(processes)

}

func ProcessParser(pid int) (Process, error) {
	var process Process

	stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return process, fmt.Errorf("read stat for PID %d: %w", pid, err)
	}

	fields := strings.Fields(string(stat))
	if len(fields) < 4 {
		return process, fmt.Errorf(
			"invalid stat for PID %d: expected at least 4 fields, got %d",
			pid,
			len(fields),
		)
	}

	parsedPID, err := strconv.Atoi(fields[0])
	if err != nil {
		return process, fmt.Errorf(
			"parse PID %q: %w",
			fields[0],
			err,
		)
	}

	ppid, err := strconv.Atoi(fields[3])
	if err != nil {
		return process, fmt.Errorf(
			"parse PPID %q for PID %d: %w",
			fields[3],
			pid,
			err,
		)
	}

	process.PID = parsedPID
	process.Name = fields[1]
	process.State = fields[2]
	process.PPID = ppid

	return process, nil
}

func PrintProcess(processes []Process) {
	w := tabwriter.NewWriter(
		os.Stdout,
		0, // minium column width
		4, // tab width
		2, // gap
		' ',
		0,
	)

	fmt.Fprintln(w, "PID\tPPID\tSTATE\tNAME")
	fmt.Fprintln(w, "---\t----\t-----\t----")

	for _, process := range processes {
		fmt.Fprintf(
			w,
			"%d\t%d\t%s\t%s\n",
			process.PID,
			process.PPID,
			process.State,
			process.Name,
		)
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "flush process table:", err)
	}

}
