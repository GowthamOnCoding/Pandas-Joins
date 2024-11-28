package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
)

func processNames() {
	// Use all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	file, err := os.Open("names.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create multiple output shards to reduce memory pressure
	const numShards = 26 // One for each letter of the alphabet
	shardMaps := make([]map[string]struct{}, numShards)
	for i := range shardMaps {
		shardMaps[i] = make(map[string]struct{}, 500000)
	}

	var wg sync.WaitGroup
	var shardMutexes [numShards]sync.Mutex

	// Super-sized channel buffer
	ch := make(chan string, 100000)

	// Use more workers than CPU cores to keep CPU busy during I/O
	numWorkers := runtime.NumCPU() * 4

	for i := 0; i < numWorkers; i++ {
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
						// Determine shard based on first letter
						shardIndex := 0
						if len(fullName) > 0 {
							shardIndex = int(fullName[0]-'a') % numShards
						}
						shardMutexes[shardIndex].Lock()
						shardMaps[shardIndex][fullName] = struct{}{}
						shardMutexes[shardIndex].Unlock()
					}
				}
			}
		}()
	}

	// Use large buffer for reading
	reader := bufio.NewReaderSize(file, 10*1024*1024) // 10MB buffer

	// Skip header
	reader.ReadString('\n')

	// Read in chunks
	chunk := make([]byte, 1024*1024) // 1MB chunks
	for {
		n, err := reader.Read(chunk)
		if err == io.EOF {
			break
		}
		ch <- string(chunk[:n])
	}
	close(ch)

	wg.Wait()

	// Merge and sort results
	totalNames := 0
	for _, shard := range shardMaps {
		totalNames += len(shard)
	}

	names := make([]string, 0, totalNames)
	for _, shard := range shardMaps {
		for name := range shard {
			names = append(names, name)
		}
	}

	sort.Strings(names)

	// Fast writing with large buffer
	outputFile, _ := os.Create("lookup.txt")
	defer outputFile.Close()

	writer := bufio.NewWriterSize(outputFile, 10*1024*1024) // 10MB buffer
	for _, name := range names {
		writer.WriteString(name + "\n")
	}
	writer.Flush()

	fmt.Println("Names written to lookup.txt successfully.")
}
