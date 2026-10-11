package main

import (
	"flag"
	"fmt"
	"os"

	"kylix/pkg/mobile"
)

func cmdMobile(args []string) {
	if len(args) == 0 {
		mobileUsage()
		os.Exit(1)
	}
	switch args[0] {
	case "doctor":
		os.Exit(mobile.Doctor(mobile.LiveToolchain(), os.Stdout))
	case "check":
		cmdMobileCheck(args[1:])
	case "serve":
		cmdMobileServe(args[1:])
	case "build":
		cmdMobileBuild(args[1:])
	case "artifact":
		cmdMobileArtifact(args[1:])
	case "help", "-h", "--help":
		mobileUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown mobile command: %s\n", args[0])
		mobileUsage()
		os.Exit(1)
	}
}

func cmdMobileCheck(args []string) {
	fs := flag.NewFlagSet("mobile check", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	backend := fs.String("backend", "go", "e2e backend: go, llvm, or both")
	hostOnly := fs.Bool("host-only", false, "Go/LLVM parity and the C ABI .so only")
	e2eOnly := fs.Bool("e2e-only", false, "curl the server only (browse, mark, review)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE: kylix mobile check [--backend=go|llvm|both] [--host-only|--e2e-only]\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if err := mobile.Check(mobile.CheckOptions{
		Backend:  *backend,
		HostOnly: *hostOnly,
		E2EOnly:  *e2eOnly,
		Out:      os.Stderr,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdMobileServe(args []string) {
	fs := flag.NewFlagSet("mobile serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	backend := fs.String("backend", "go", "server backend: go or llvm")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE: kylix mobile serve [--backend=go|llvm]\n\nListens on :8091 in the foreground until Ctrl+C.\nThe server prints \"vocab server on :8091\".\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if err := mobile.Serve("", *backend); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdMobileBuild(args []string) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "USAGE: kylix mobile build android amd64|arm64\n       kylix mobile build ios simulator|device\n")
		os.Exit(2)
	}
	var err error
	switch args[0] {
	case "android":
		err = mobile.BuildAndroid("", args[1], os.Stderr)
	case "ios":
		err = mobile.BuildIOS("", args[1], os.Stderr)
	default:
		fmt.Fprintf(os.Stderr, "USAGE: kylix mobile build android amd64|arm64\n       kylix mobile build ios simulator|device\n")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdMobileArtifact(args []string) {
	if len(args) != 3 {
		fmt.Fprintf(os.Stderr, "USAGE: kylix mobile artifact android <libkylixvocab.so> amd64|arm64\n       kylix mobile artifact ios <libkylixvocab.a> simulator|device\n")
		os.Exit(2)
	}
	var err error
	switch args[0] {
	case "android":
		err = mobile.ArtifactAndroid("", args[1], args[2], os.Stderr)
	case "ios":
		err = mobile.ArtifactIOS("", args[1], args[2], os.Stderr)
	default:
		fmt.Fprintf(os.Stderr, "USAGE: kylix mobile artifact android <so> amd64|arm64\n       kylix mobile artifact ios <archive> simulator|device\n")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func mobileUsage() {
	fmt.Printf(`USAGE: kylix mobile <command>

Flashcard (apps/vocab) commands. One implementation, shared with the .sh scripts.

COMMANDS:
    doctor                 Required tools, Android NDK, macOS-only iOS, libs this example does not use
    check                  Go/LLVM parity, C ABI, and curl browse / mark / review
    serve                  Leave the vocab server in the foreground on :8091
    build android amd64    libkylixvocab.so for an x86_64 emulator (jniLibs/x86_64)
    build android arm64    libkylixvocab.so for an arm64 emulator (jniLibs/arm64-v8a)
    build ios simulator    libkylixvocab.a (macOS + Xcode + xcodegen; refused elsewhere)
    build ios device       libkylixvocab.a for a device (macOS only)
    artifact android|ios   ELF/Mach-O and export check for an existing library

EXAMPLES:
    kylix mobile doctor
    kylix mobile check
    kylix mobile check --e2e-only --backend=llvm
    kylix mobile serve
    kylix mobile build android amd64
`)
}
