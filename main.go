package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gadumitrachioaiei/go-lox/scanner"
)

func main() {
	if args := os.Args; len(args) > 2 {
		log.Fatal("We need at most one argument, that must be a file path")
	} else if len(args) == 2 {
		runFile(args[1])
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("reading file: %v", err)
	}
	run(string(data))
}

func runPrompt() {
	ioScanner := bufio.NewScanner(os.Stdin)
	for ioScanner.Scan() {
		run(ioScanner.Text())
	}
	if err := ioScanner.Err(); err != nil {
		log.Fatalf("scanning stdin: %v", err)
	}
}

func run(text string) {
	scanner := scanner.New(text)
	tokens, errors := scanner.ScanTokens()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println(err)
		}
		return
	}
	for _, token := range tokens {
		fmt.Println(token)
	}
}
