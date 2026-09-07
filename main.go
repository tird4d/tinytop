package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
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

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "list":
			ProcessList()

		case "stop", "terminate":
			KillProcess(parts)

		case "quit", "exit":
			return

		default:
			fmt.Println("unknown command:", parts[0])
		}

	}

}

func RefreshProcessList() {
	for {
		ProcessList()
		time.Sleep(5 * time.Second)
	}

}

func ProcessList() {
	ps, _ := os.ReadDir("/proc")
	var processes []Process

	for _, p := range ps {
		fmt.Println(p)
		r, err := strconv.Atoi(p.Name())
		if err != nil {
			continue
		}

		processId := r

		MyProcess, err := ProcessParser(processId)
		if err != nil {
			// fmt.Println(err.Error())
			continue
		}

		processes = append(processes, MyProcess)

	}

	slices.SortFunc(processes, func(a, b Process) int {
		switch {
		case a.PID > b.PID:
			return -1
		case a.PID < b.PID:
			return 1
		default:
			return 0
		}
	})

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

func KillProcess(parts []string) {
	if len(parts) != 2 {
		fmt.Printf("usage: %s <pid>\n", parts[0])
		return
	}

	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		fmt.Println("invalid PID:", parts[1])
		return
	}
	p, err := os.FindProcess(pid)
	err = p.Kill()

	if err != nil {
		fmt.Printf("Cannot kill the process %s\n", err.Error())
	}

	fmt.Printf("%s requested for PID %d\n", parts[0], pid)

}
