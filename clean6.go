package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type ColumnInfo struct {
	index        int
	name         string
	isNameColumn bool
}

func getColumnIndices(headerLine string) []ColumnInfo {
	headers := strings.Split(headerLine, ",")
	columns := make([]ColumnInfo, 0)

	for idx, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		isNameCol := strings.Contains(header, "name") || 
			         strings.Contains(header, "first") || 
					 strings.Contains(header, "last") ||
					 strings.Contains(header, "given")
		
		columns = append(columns, ColumnInfo{
			index:        idx,
			name:         header,
			isNameColumn: isNameCol,
		})
	}

	fmt.Printf("Detected columns:\n")
	for _, col := range columns {
		fmt.Printf("- %s (index: %d, name column: %v)\n", 
			col.name, col.index, col.isNameColumn)
	}

	return columns
}

func processNames() {
	startTime := time.Now()
	fmt.Println("Starting dynamic column processing job...")
	
	runtime.GOMAXPROCS(runtime.NumCPU())

	file, err := os.Open("names.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create scanner with buffer configuration
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 10*1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	// Read header
	if !scanner.Scan() {
		fmt.Println("Error reading header:", scanner.Err())
		return
	}
	headerLine := scanner.Text()
	columns := getColumnIndices(headerLine)

	var names []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	var totalInputRows int64

	ch := make(chan string, 100000)
	numWorkers := runtime.NumCPU() * 4

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localNames := make([]string, 0)

			for line := range ch {
				fields := strings.Split(line, ",")
				if len(fields) >= len(columns) {
					nameComponents := make([]string, 0)
					
					for _, col := range columns {
						if col.isNameColumn {
							value := strings.TrimSpace(fields[col.index])
							if value != "" {
								nameComponents = append(nameComponents, value)
							}
						}
					}

					if len(nameComponents) > 0 {
						fullName := strings.ToLower(strings.Join(nameComponents, " "))
						if len(fullName) > 2 {
							localNames = append(localNames, fullName)
						}
					}
				}
			}

			mu.Lock()
			names = append(names, localNames...)
			mu.Unlock()
		}()
	}

	scanStart := time.Now()
	for scanner.Scan() {
		ch <- scanner.Text()
		totalInputRows++
		if totalInputRows%1000000 == 0 {
			fmt.Printf("Processed %d million rows (Speed: %.2f rows/sec)\n", 
				totalInputRows/1000000, float64(totalInputRows)/time.Since(scanStart).Seconds())
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error during scanning:", err)
	}

	close(ch)
	wg.Wait()

	fmt.Println("Sorting names...")
	sort.Strings(names)

	outputFile, _ := os.Create("lookup.txt")
	defer outputFile.Close()

	writer := bufio.NewWriterSize(outputFile, 10*1024*1024)
	for _, name := range names {
		writer.WriteString(name + "\n")
	}
	writer.Flush()

	fmt.Printf("\nProcessing Summary:\n"+
		"Input rows processed: %d\n"+
		"Names generated: %d\n"+
		"Total time: %s\n",
		totalInputRows,
		len(names),
		time.Since(startTime))
}
