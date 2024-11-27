package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/schollz/progressbar/v3"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

type Config struct {
	WorkerCount int `json:"worker_count"`
	BatchSize   int `json:"batch_size"`
	BufferSize  int `json:"buffer_size"`
}

type Metrics struct {
	ProcessedRecords int64         `json:"processed_records"`
	MatchedRecords   int64         `json:"matched_records"`
	ProcessingTime   time.Duration `json:"processing_time"`
}

type Record struct {
	Name         string `csv:"name"`
	Result       bool   `csv:"lookup_result"`
	MatchedValue string `csv:"matched_value"`
	MatchType    string `csv:"match_type"`
}

type CacheEntry struct {
	matchedValue string
	matchType    string
}

type CacheStats struct {
	hits   uint64
	misses uint64
}

type Server struct {
	db         *leveldb.DB
	matcher    *search.Matcher
	config     Config
	matchCache sync.Map
	cacheStats CacheStats
}

type StringLookupRequest struct {
	SearchString string `json:"search_string"`
}

type StringLookupResponse struct {
	Found        bool   `json:"found"`
	MatchedValue string `json:"matched_value"`
	MatchType    string `json:"match_type"`
	CacheHit     bool   `json:"cache_hit"`
}

type FileProcessRequest struct {
	InputFilePath string   `json:"input_file_path"`
	SearchColumns []string `json:"search_columns"`
}

type FileProcessResponse struct {
	Metrics       *Metrics `json:"metrics"`
	ProcessedPath string   `json:"processed_path"`
	CacheStats    struct {
		Hits   uint64 `json:"cache_hits"`
		Misses uint64 `json:"cache_misses"`
	} `json:"cache_stats"`
}

func NewServer(lookupFile string, config Config) (*Server, error) {
	db, err := leveldb.OpenFile("lookup.db", nil)
	if err != nil {
		return nil, err
	}

	server := &Server{
		db:      db,
		matcher: search.New(language.English, search.Loose),
		config:  config,
	}

	if err := server.loadLookupData(lookupFile); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) loadLookupData(lookupFile string) error {
	file, err := os.Open(lookupFile)
	if err != nil {
		return err
	}
	defer file.Close()

	batch := new(leveldb.Batch)
	scanner := bufio.NewScanner(file)
	bar := progressbar.Default(-1, "Loading Lookup Data")

	for scanner.Scan() {
		value := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if len(value) > 0 {
			prefixLen := min(3, len(value))
			prefix := value[:prefixLen]
			key := prefix + ":" + value
			batch.Put([]byte(key), []byte{1})
		}
		bar.Add(1)
	}

	return s.db.Write(batch, nil)
}

func (s *Server) lookupWithCache(searchValue string) (bool, string, string) {
	if entry, ok := s.matchCache.Load(searchValue); ok {
		atomic.AddUint64(&s.cacheStats.hits, 1)
		cacheEntry := entry.(CacheEntry)
		return true, cacheEntry.matchedValue, cacheEntry.matchType
	}

	atomic.AddUint64(&s.cacheStats.misses, 1)
	found, matchedValue, matchType := s.performLookup(searchValue)

	if found {
		s.matchCache.Store(searchValue, CacheEntry{
			matchedValue: matchedValue,
			matchType:    matchType,
		})
	}

	return found, matchedValue, matchType
}

func (s *Server) performLookup(searchValue string) (bool, string, string) {
	searchValue = strings.ToLower(searchValue)
	if len(searchValue) == 0 {
		return false, "", ""
	}

	prefixLen := min(3, len(searchValue))
	prefix := searchValue[:prefixLen]

	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix+":")), nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		lookupValue := strings.TrimPrefix(key, prefix+":")

		if start, _ := s.matcher.IndexString(searchValue, lookupValue); start != -1 {
			return true, lookupValue, "contains_lookup"
		}
		if start, _ := s.matcher.IndexString(lookupValue, searchValue); start != -1 {
			return true, lookupValue, "lookup_contains"
		}
	}

	return false, "", ""
}

func (s *Server) processInputFile(inputFile string, searchColumns []string) (*Metrics, error) {
	startTime := time.Now()
	metrics := &Metrics{}
	tmpFile := inputFile + ".tmp"

	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return metrics, err
	}

	var records []Record
	if err := csvutil.Unmarshal(inputData, &records); err != nil {
		return metrics, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})

	bar := progressbar.Default(int64(len(records)), "Processing Records")

	var wg sync.WaitGroup
	recordsChan := make(chan []Record, s.config.BufferSize)
	resultsChan := make(chan []Record, s.config.BufferSize)

	for i := 0; i < s.config.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range recordsChan {
				processedBatch := make([]Record, len(batch))
				copy(processedBatch, batch)

				for i := range processedBatch {
					found, matchedValue, matchType := s.lookupWithCache(strings.ToLower(processedBatch[i].Name))
					processedBatch[i].Result = found
					processedBatch[i].MatchedValue = matchedValue
					processedBatch[i].MatchType = matchType

					if found {
						atomic.AddInt64(&metrics.MatchedRecords, 1)
					}
					atomic.AddInt64(&metrics.ProcessedRecords, 1)
					bar.Add(1)
				}
				resultsChan <- processedBatch
			}
		}()
	}

	go func() {
		batchSize := s.config.BatchSize
		for i := 0; i < len(records); i += batchSize {
			end := i + batchSize
			if end > len(records) {
				end = len(records)
			}
			recordsChan <- records[i:end]
		}
		close(recordsChan)
	}()

	results := make([]Record, 0, len(records))
	go func() {
		for batch := range resultsChan {
			results = append(results, batch...)
		}
	}()

	wg.Wait()
	close(resultsChan)

	output, err := os.Create(tmpFile)
	if err != nil {
		return metrics, err
	}
	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	if err := writer.Write([]string{"name", "lookup_result", "matched_value", "match_type"}); err != nil {
		return metrics, err
	}

	for _, record := range results {
		if err := writer.Write([]string{
			record.Name,
			strconv.FormatBool(record.Result),
			record.MatchedValue,
			record.MatchType,
		}); err != nil {
			return metrics, err
		}
	}

	metrics.ProcessingTime = time.Since(startTime)

	if err := os.Rename(tmpFile, inputFile); err != nil {
		return metrics, err
	}

	return metrics, nil
}

func (s *Server) handleStringLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StringLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheHitsBefore := atomic.LoadUint64(&s.cacheStats.hits)
	found, matchedValue, matchType := s.lookupWithCache(strings.ToLower(req.SearchString))
	cacheHit := atomic.LoadUint64(&s.cacheStats.hits) > cacheHitsBefore

	response := StringLookupResponse{
		Found:        found,
		MatchedValue: matchedValue,
		MatchType:    matchType,
		CacheHit:     cacheHit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleFileProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FileProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metrics, err := s.processInputFile(req.InputFilePath, req.SearchColumns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := FileProcessResponse{
		Metrics:       metrics,
		ProcessedPath: req.InputFilePath,
		CacheStats: struct {
			Hits   uint64 `json:"cache_hits"`
			Misses uint64 `json:"cache_misses"`
		}{
			Hits:   atomic.LoadUint64(&s.cacheStats.hits),
			Misses: atomic.LoadUint64(&s.cacheStats.misses),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	inputFile := flag.String("input", "", "Input CSV file path")
	lookupFile := flag.String("lookup", "", "Lookup file path")
	workers := flag.Int("workers", 4, "Number of workers")
	batchSize := flag.Int("batch", 1000, "Batch size")
	bufferSize := flag.Int("buffer", 100, "Buffer size")
	port := flag.String("port", "", "Port for API server")
	flag.Parse()

	config := Config{
		WorkerCount: *workers,
		BatchSize:   *batchSize,
		BufferSize:  *bufferSize,
	}

	if *port != "" {
		server, err := NewServer(*lookupFile, config)
		if err != nil {
			log.Fatal(err)
		}
		defer server.db.Close()

		http.HandleFunc("/lookup", server.handleStringLookup)
		http.HandleFunc("/process-file", server.handleFileProcess)

		log.Printf("Server starting on port %s", *port)
		log.Fatal(http.ListenAndServe(":"+*port, nil))
		return
	}

	if *inputFile == "" || *lookupFile == "" {
		flag.Usage()
		return
	}

	server, err := NewServer(*lookupFile, config)
	if err != nil {
		log.Fatal(err)
	}
	defer server.db.Close()

	metrics, err := server.processInputFile(*inputFile, []string{"name"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Processing completed:\n")
	fmt.Printf("Total records processed: %d\n", metrics.ProcessedRecords)
	fmt.Printf("Matched records: %d\n", metrics.MatchedRecords)
	fmt.Printf("Processing time: %v\n", metrics.ProcessingTime)
	fmt.Printf("Cache hits: %d\n", atomic.LoadUint64(&server.cacheStats.hits))
	fmt.Printf("Cache misses: %d\n", atomic.LoadUint64(&server.cacheStats.misses))
}
