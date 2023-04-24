package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/unix"
)

const FLAG = "CTF{you_made_it_past_the_hurdles}"

type ShurdleTest func() error

var shurdles = []ShurdleTest{
	RequireAtLeastOneArgument,
	RequireParentProcessExe("/usr/bin/bash"),
	ExpectProcessName("/shurdles"),
	ExpectEnvironmentVariable("HACKERS", "hack\nthe\nplanet", "hack the planet on separate lines"),
	RequireNoLDPreload,
	RequireWorkdir("/run/. -- !!"),
	RequireLeetFD,
	RequireTZ("America/Los_Angeles"),
	RequireShurdlesHelper,
	RequireOldShurdlesCache,
}

func main() {
	for i := 0; i < len(shurdles); i++ {
		if err := shurdles[i](); err != nil {
			fmt.Printf("shurdle %d failed: %v\n", i, err)
			os.Exit(1)
		}
	}
	fmt.Printf("Congratulations!!!\n")
	fmt.Printf("%s\n", FLAG)
}

func RequireAtLeastOneArgument() error {
	if len(os.Args) < 2 {
		return errors.New("expected at least 1 argument")
	}
	return nil
}

func ExpectProcessName(name string) func() error {
	return func() error {
		if len(os.Args) < 1 {
			return errors.New("no process name given")
		}
		if os.Args[0] != name {
			return fmt.Errorf("I expected to be called %s, not %s", name, os.Args[0])
		}
		return nil
	}
}

func ExpectEnvironmentVariable(name, value, desc string) func() error {
	return func() error {
		if os.Getenv(name) != value {
			return fmt.Errorf("I expected the environment variable %q to look like %s", name, desc)
		}
		return nil
	}
}

func RequireNoLDPreload() error {
	if _, present := os.LookupEnv("LD_PRELOAD"); present {
		return errors.New("please don't try to LD_PRELOAD me")
	}
	return nil
}

func RequireLeetFD() error {
	if !isFdValid(3) {
		return errors.New("fd 3 isn't open")
	}
	f := os.NewFile(uintptr(3), "-")
	if f == nil {
		return errors.New("fd 3 isn't valid")
	}
	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("fd 3 stat err: %w", err)
	}
	if fi.Size() != 1337 {
		return fmt.Errorf("expected fd 3 to be a file of 1337 bytes in length")
	}
	return nil
}

func isFdValid(fd int) bool {
	_, err := unix.FcntlInt(uintptr(fd), unix.F_GETFD, 0)
	return err == nil
}

func RequireWorkdir(path string) func() error {
	return func() error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("workdir error: %w", err)
		}
		if dir != path {
			return fmt.Errorf("expected workdir %q", path)
		}
		return nil
	}
}

func RequireTZ(tzname string) func() error {
	return func() error {
		if time.Local.String() != tzname {
			return fmt.Errorf("tz %s != %s", time.Local.String(), tzname)
			//return errors.New("expected to be run in the same timezone as BSidesSF")
		}
		return nil
	}
}

func RequireParentProcessExe(name string) func() error {
	return func() error {
		ppid := os.Getppid()
		exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", ppid))
		if err != nil {
			return fmt.Errorf("could not get parent process: %w", err)
		}
		if exe != name {
			return fmt.Errorf("expected to be executed by %q, not %q", name, exe)
		}
		return nil
	}
}

func RequireShurdlesHelper() error {
	path, err := exec.LookPath("shurdles-helper")
	if err != nil {
		return errors.New("could not find shurdles-helper")
	}
	cmd := exec.Command(path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shurdles-helper failed: %w", err)
	}
	return nil
}

func RequireOldShurdlesCache() error {
	fname := "/home/ctf/.cache/shurdles"
	fi, err := os.Stat(fname)
	if err != nil {
		return fmt.Errorf("expected %s, does it exist?", fname)
	}
	dayAgo := time.Now().Add(-24 * time.Hour)
	mtime := fi.ModTime()
	if mtime.After(dayAgo) {
		return fmt.Errorf("%s was modified in the last day, sorry", fname)
	}
	return nil
}
