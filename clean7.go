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

// ColumnInfo stores metadata about each CSV column
type ColumnInfo struct {
	index        int
	name         string
	isNameColumn bool
}

// cleanName removes non-alphabetic characters and extra spaces
func cleanName(name string) string {
	// Keep only letters and spaces
	var result strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	// Remove extra spaces and trim
	reg := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(reg.ReplaceAllString(result.String(), " "))
}

func processNames() {
	startTime := time.Now()
	fmt.Println("Starting enhanced name processing job...")
	
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("Using %d CPU cores\n", runtime.NumCPU())

	// Open input file
	file, err := os.Open("names.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Initialize scanner with large buffer
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	// Read and process header
	if !scanner.Scan() {
		fmt.Println("Error reading header:", scanner.Err())
		return
	}
	headers := strings.Split(scanner.Text(), ",")
	fmt.Printf("Found %d columns in CSV\n", len(headers))

	var allNames []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	var totalInputRows, totalNamesFound int64

	ch := make(chan string, 100000)
	numWorkers := runtime.NumCPU() * 4
	fmt.Printf("Launching %d worker goroutines\n", numWorkers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localNames := make([]string, 0)

			for line := range ch {
				fields := strings.Split(line, ",")
				// Process each field as a potential name
				for _, field := range fields {
					cleanedName := cleanName(strings.ToLower(field))
					if len(cleanedName) >= 3 {
						localNames = append(localNames, cleanedName)
					}
				}
			}

			mu.Lock()
			allNames = append(allNames, localNames...)
			totalNamesFound += int64(len(localNames))
			mu.Unlock()
		}()
	}

	// Process input file
	fmt.Println("Starting file processing...")
	scanStart := time.Now()
	for scanner.Scan() {
		ch <- scanner.Text()
		totalInputRows++
		if totalInputRows%1000000 == 0 {
			fmt.Printf("Processed %d million rows (Speed: %.2f rows/sec)\n", 
				totalInputRows/1000000, float64(totalInputRows)/time.Since(scanStart).Seconds())
		}
	}

	close(ch)
	fmt.Println("Waiting for workers to complete...")
	wg.Wait()

	// Sort all names
	fmt.Printf("Sorting %d names...\n", len(allNames))
	sortStart := time.Now()
	sort.Strings(allNames)
	fmt.Printf("Sorting completed in %s\n", time.Since(sortStart))

	// Write all sorted names
	fmt.Println("Writing sorted names to file...")
	writeStart := time.Now()
	allNamesFile, _ := os.Create("all_names.txt")
	defer allNamesFile.Close()
	writer := bufio.NewWriterSize(allNamesFile, 10*1024*1024)
	for _, name := range allNames {
		writer.WriteString(name + "\n")
	}
	writer.Flush()

	// Remove duplicates and write unique names
	fmt.Println("Removing duplicates...")
	uniqueStart := time.Now()
	uniqueNames := make([]string, 0)
	seen := make(map[string]struct{})
	for _, name := range allNames {
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			uniqueNames = append(uniqueNames, name)
		}
	}

	uniqueFile, _ := os.Create("unique_names.txt")
	defer uniqueFile.Close()
	uniqueWriter := bufio.NewWriterSize(uniqueFile, 10*1024*1024)
	for _, name := range uniqueNames {
		uniqueWriter.WriteString(name + "\n")
	}
	uniqueWriter.Flush()

	// Print final statistics
	fmt.Printf("\nProcessing Summary:\n"+
		"Total time: %s\n"+
		"Input rows processed: %d\n"+
		"Total names found: %d\n"+
		"Unique names: %d\n"+
		"Duplicate names removed: %d\n"+
		"Processing speed: %.2f rows/sec\n"+
		"Memory used: %.2f MB\n",
		time.Since(startTime),
		totalInputRows,
		totalNamesFound,
		len(uniqueNames),
		len(allNames)-len(uniqueNames),
		float64(totalInputRows)/time.Since(startTime).Seconds(),
		float64(runtime.MemStats{}.Alloc)/(1024*1024))
}
