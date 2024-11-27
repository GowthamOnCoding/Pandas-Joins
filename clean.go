package main

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strings"
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
    scanner := bufio.NewScanner(file)

    // Skip the header line
    if scanner.Scan() {
        // Read header
    }

    for scanner.Scan() {
        line := scanner.Text()
        columns := strings.Split(line, ",")
        if len(columns) >= 5 {
            givenNameOne := strings.TrimSpace(columns[0])
            lastName := strings.TrimSpace(columns[4])
            if givenNameOne != "" && lastName != "" {
                fullName := strings.ToLower(givenNameOne + " " + lastName)
                namesMap[fullName] = struct{}{}
            }
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
        return
    }

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
