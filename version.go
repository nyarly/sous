package main

// These values are set at build time using -ldflags "-X main.Name=Value"
var Version, Branch, CommitSHA, Revision, BuildNumber, BuildTimestamp, OS, Arch string

func init() {
	Revision = CommitSHA
}
