package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "regexp"
    "strings"
    "sync"
    "github.com/dgraph-io/ristretto"
    "golang.org/x/sync/errgroup"
)

type Cache struct {
    cache *ristretto.Cache
}

func NewCache(maxCost int64) (*Cache, error) {
    cache, err := ristretto.NewCache(&ristretto.Config{
        NumCounters: 1e7,     // number of keys to track frequency of (10M).
        MaxCost:     maxCost, // maximum cost of cache (in bytes).
        BufferItems: 64,      // number of keys per Get buffer.
    })
    if err != nil {
        return nil, err
    }
    return &Cache{cache: cache}, nil
}

func (c *Cache) Get(key string) ([]int, bool) {
    val, found := c.cache.Get(key)
    if found {
        return val.([]int), true
    }
    return nil, false
}

func (c *Cache) Set(key string, value []int) {
    c.cache.Set(key, value, int64(len(value)))
    c.cache.Wait()
}

func normalizeString(s string) string {
    re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
    return strings.ToLower(re.ReplaceAllString(s, ""))
}

func processChunk(ctx context.Context, filePath string, offset, chunkSize int64, searchStrings []string, cache *Cache, results map[string][]int, mu *sync.Mutex, resultChan chan<- string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    file.Seek(offset, 0)
    scanner := bufio.NewScanner(file)
    scanner.Buffer(make([]byte, chunkSize), int(chunkSize))

    lineNum := int(offset / chunkSize)
    for scanner.Scan() {
        line := scanner.Text()
        lineNum++
        normalizedLine := normalizeString(line)
        for _, str := range searchStrings {
            normalizedStr := normalizeString(str)
            if strings.Contains(normalizedStr, normalizedLine) {
                mu.Lock()
                results[str] = append(results[str], lineNum)
                mu.Unlock()
                resultChan <- fmt.Sprintf("Search String: '%s', Value: '%s', Line Number: %d\n", str, line, lineNum)
                cache.Set(str, results[str])
            }
        }
    }

    return scanner.Err()
}

func searchInFile(ctx context.Context, filePath string, searchStrings []string, cache *Cache, chunkSize int64, resultFilePath string) (map[string][]int, error) {
    file, err := os.Stat(filePath)
    if err != nil {
        return nil, err
    }

    fileSize := file.Size()
    results := make(map[string][]int)
    for _, str := range searchStrings {
        if cachedResult, found := cache.Get(str); found {
            results[str] = cachedResult
        } else {
            results[str] = []int{}
        }
    }

    resultFile, err := os.Create(resultFilePath)
    if err != nil {
        return nil, err
    }
    defer resultFile.Close()

    resultChan := make(chan string, 100)
    var wg sync.WaitGroup

    // Goroutine for writing results to file
    wg.Add(1)
    go func() {
        defer wg.Done()
        for result := range resultChan {
            resultFile.WriteString(result)
        }
    }()

    g, ctx := errgroup.WithContext(ctx)
    mu := &sync.Mutex{}

    for offset := int64(0); offset < fileSize; offset += chunkSize {
        offset := offset // capture the current offset
        g.Go(func() error {
            return processChunk(ctx, filePath, offset, chunkSize, searchStrings, cache, results, mu, resultChan)
        })
    }

    if err := g.Wait(); err != nil {
        close(resultChan)
        return nil, err
    }

    close(resultChan)
    wg.Wait()

    for str, lines := range results {
        cache.Set(str, lines)
    }

    return results, nil
}

func main() {
    searchStrings := []string{"error", "warning", "info"}
    filePath := "largefile.txt"
    cache, err := NewCache(1 << 30) // 1GB cache
    if err != nil {
        fmt.Println("Error creating cache:", err)
        return
    }
    chunkSize := int64(1024 * 1024) // 1MB chunks
    resultFilePath := "search_results.txt"

    ctx := context.Background()
    results, err := searchInFile(ctx, filePath, searchStrings, cache, chunkSize, resultFilePath)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    for str, lines := range results {
        fmt.Printf("String '%s' found in lines: %v\n", str, lines)
    }
}
