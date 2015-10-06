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
	fmt.Fprintf(os.Stderr, ".")
}

func (p Progress) Done(message string) {
	fmt.Fprintf(os.Stderr, message+"\n")
}
