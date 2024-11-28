package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

func main() {
	inputFileName := "names.csv" // Replace with your input file name

	file, err := os.Open(inputFileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	namesMap := make(map[string]struct{})
	var mu sync.Mutex
	var wg sync.WaitGroup

	ch := make(chan string, 100)

	// Start goroutine to process lines
	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range ch {
			columns := strings.Split(line, ",")
			if len(columns) >= 5 {
				givenNameOne := strings.TrimSpace(columns[0])
				lastName := strings.TrimSpace(columns[4])
				if givenNameOne != "" && lastName != "" {
					fullName := strings.ToLower(givenNameOne + " " + lastName)
					mu.Lock()
					namesMap[fullName] = struct{}{}
					mu.Unlock()
				}
			}
		}
	}()

	reader := bufio.NewReader(file)

	// Skip the header line
	_, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != os.EOF {
				fmt.Println("Error reading file:", err)
			}
			break
		}
		ch <- line
	}
	close(ch)

	// Wait for processing goroutine to finish
	wg.Wait()

	names := make([]string, 0, len(namesMap))
	for name := range namesMap {
		names = append(names, name)
	}

	sort.Strings(names)

	outputFile, err := os.Create("lookup.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, name := range names {
		writer.WriteString(name + "\n")
	}
	writer.Flush()

	fmt.Println("Names written to lookup.txt successfully.")
}
