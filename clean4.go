package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

func isValidName(name string) bool {
	if len(name) <= 2 {
		return false
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func processNames() {
	startTime := time.Now()
	fmt.Println("Starting name processing job...")

	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Using %d CPU cores\n", runtime.NumCPU())

	file, err := os.Open("names.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	fmt.Printf("Successfully opened input file at: %s\n", time.Since(startTime))

	const numShards = 26
	shardMaps := make([]map[string]struct{}, numShards)
	for i := range shardMaps {
		shardMaps[i] = make(map[string]struct{}, 1000000)
	}
	fmt.Printf("Initialized %d shards for data processing\n", numShards)

	var wg sync.WaitGroup
	var shardMutexes [numShards]sync.Mutex
	var processedLines uint64
	var validNames uint64
	var mu sync.Mutex

	var totalInputRows int64

	ch := make(chan string, 100000)
	numWorkers := runtime.NumCPU() * 4
	fmt.Printf("Launching %d worker goroutines\n", numWorkers)

	csvSplitter := regexp.MustCompile(`[^,]*,|"[^"]*"`)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			localProcessed := uint64(0)
			localValid := uint64(0)

			for line := range ch {
				localProcessed++
				matches := csvSplitter.FindAllString(line, -1)
				if len(matches) >= 5 {
					givenNameOne := strings.Trim(strings.Trim(matches[0], ","), "\"")
					lastName := strings.Trim(strings.Trim(matches[4], ","), "\"")

					if givenNameOne != "" && lastName != "" {
						fullName := strings.ToLower(strings.TrimSpace(givenNameOne + " " + lastName))

						if isValidName(fullName) {
							localValid++
							shardIndex := int(fullName[0]-'a') % numShards
							if shardIndex >= 0 && shardIndex < numShards {
								shardMutexes[shardIndex].Lock()
								shardMaps[shardIndex][fullName] = struct{}{}
								shardMutexes[shardIndex].Unlock()
							}
						}
					}
				}

				if localProcessed%100000 == 0 {
					mu.Lock()
					processedLines += localProcessed
					validNames += localValid
					localProcessed = 0
					localValid = 0
					fmt.Printf("Progress: Processed %d lines, Found %d valid names (Time: %s)\n",
						processedLines, validNames, time.Since(startTime))
					mu.Unlock()
				}
			}

			mu.Lock()
			processedLines += localProcessed
			validNames += localValid
			mu.Unlock()
		}(i)
	}

	fmt.Println("Starting file scanning...")
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	totalInputRows++
	scanner.Scan() // Skip header

	scanStart := time.Now()
	for scanner.Scan() {
		ch <- scanner.Text()
		totalInputRows++
		if totalInputRows%1000000 == 0 {
			fmt.Printf("Scanned %d million rows (Speed: %.2f rows/sec)\n",
				totalInputRows/1000000, float64(totalInputRows)/time.Since(scanStart).Seconds())
		}
	}
	close(ch)
	fmt.Printf("File scanning complete. Total rows: %d (Time: %s)\n", totalInputRows, time.Since(scanStart))

	fmt.Println("Waiting for workers to finish processing...")
	wg.Wait()
	processingTime := time.Since(startTime)
	fmt.Printf("Processing complete. Time taken: %s\n", processingTime)

	fmt.Println("Merging and sorting results...")
	mergeStart := time.Now()
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
	fmt.Printf("Merge and sort complete. Time taken: %s\n", time.Since(mergeStart))

	fmt.Println("Writing results to file...")
	writeStart := time.Now()
	outputFile, _ := os.Create("lookup.txt")
	defer outputFile.Close()

	writer := bufio.NewWriterSize(outputFile, 10*1024*1024)
	for _, name := range names {
		writer.WriteString(name + "\n")
	}
	writer.Flush()

	fmt.Printf("\nFinal Statistics:\n"+
		"Total time: %s\n"+
		"Input rows scanned: %d\n"+
		"Total lines processed: %d\n"+
		"Processing speed: %.2f lines/sec\n"+
		"Valid unique names found: %d\n"+
		"Output file write time: %s\n",
		time.Since(startTime),
		totalInputRows,
		processedLines,
		float64(processedLines)/processingTime.Seconds(),
		len(names),
		time.Since(writeStart))
}
