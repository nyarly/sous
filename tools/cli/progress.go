package cli

import (
	"fmt"
	"os"
)

type Progress string

func BeginProgress(title string) Progress {
	fmt.Fprintf(os.Stderr, title+"..")
	return Progress(title)
}

func (p Progress) Increment() {
	if p != "" {
		fmt.Fprintf(os.Stderr, ".")
		return
	}
}

func (p Progress) Done(message string) {
	if p != "" {
		fmt.Fprintf(os.Stderr, message+"\n")
	}
}
