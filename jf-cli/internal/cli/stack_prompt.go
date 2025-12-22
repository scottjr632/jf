package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/scottjr632/jf-cli/internal/stack"
)

func promptNextCommitSelection(commits []stack.StackItem) (stack.StackItem, error) {
	if len(commits) == 0 {
		return stack.StackItem{}, fmt.Errorf("no commits to select")
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintln(os.Stdout, "Select next commit:")
		for i, item := range commits {
			fmt.Fprintf(os.Stdout, "  %d) %s %s\n", i+1, item.Commit.Short, item.Commit.Subject)
		}
		fmt.Fprint(os.Stdout, "Enter number: ")

		text, err := reader.ReadString('\n')
		if err != nil && len(text) == 0 {
			return stack.StackItem{}, err
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			if err == io.EOF {
				return stack.StackItem{}, err
			}
			continue
		}
		index, convErr := strconv.Atoi(trimmed)
		if convErr != nil || index < 1 || index > len(commits) {
			fmt.Fprintln(os.Stdout, "Invalid selection.")
			if err == io.EOF {
				return stack.StackItem{}, err
			}
			continue
		}
		return commits[index-1], nil
	}
}
