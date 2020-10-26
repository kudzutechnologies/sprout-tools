package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type CSVDataFile struct {
	separator       string
	filename        string
	headers         []string
	headersToAppend []string
	hasData         bool
}

func CreateCSVDataFile(path string, separator string) (*CSVDataFile, error) {
	var (
		headers         []string = nil
		headersToAppend []string = nil
		hasData         bool     = false
	)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s: %w", path, err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		firstLine := scanner.Text()
		headers = strings.Split(firstLine, separator)

		// Check if there are more lines after the header
		hasData = scanner.Scan()
	}

	// If there are no data, add default headers
	if !hasData {
		headersToAppend = append(headersToAppend, "Timestamp")
	}

	// if err := scanner.Err(); err != nil {
	// 	return nil, fmt.Errorf("Could not read header line from %s: %w", path, err)
	// }

	return &CSVDataFile{
		separator:       separator,
		filename:        path,
		headers:         headers,
		headersToAppend: headersToAppend,
		hasData:         hasData,
	}, nil
}

func (csv *CSVDataFile) writeHeaders() error {
	if csv.headersToAppend == nil {
		return nil
	}

	// Create in a temporary file where we are going to write everything
	tmpFile := fmt.Sprintf("%s.tmp", csv.filename)
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("Could not create temporary file %s: %w", tmpFile, err)
	}

	// Start by writing the new headers
	var allHeaders []string = nil
	for _, hdr := range csv.headers {
		allHeaders = append(allHeaders, hdr)
	}
	for _, hdr := range csv.headersToAppend {
		allHeaders = append(allHeaders, hdr)
	}

	hdrLine := []byte(strings.Join(allHeaders, csv.separator))
	hdrLine = append(hdrLine, '\n')
	_, err = f.Write(hdrLine)
	if err != nil {
		f.Close()
		return fmt.Errorf("Error writing to %s: %w", csv.filename, err)
	}

	// If we have data, copy the rest of the file contents
	if csv.hasData {
		fIn, err := os.Open(csv.filename)
		if err != nil {
			return fmt.Errorf("Could not open %s for reading: %w", csv.filename, err)
		}

		scanner := bufio.NewScanner(fIn)
		if !scanner.Scan() { // Skip header
			fIn.Close()
			return fmt.Errorf("Could not scan header of %s", csv.filename)
		}

		for scanner.Scan() {
			bt := scanner.Bytes()
			bt = append(bt, '\n')
			_, err := f.Write(bt)
			if err != nil {
				fIn.Close()
				f.Close()
				return fmt.Errorf("Error writing to %s: %w", csv.filename, err)
			}
		}
		if err := scanner.Err(); err != nil {
			fIn.Close()
			return fmt.Errorf("Error reading %s: %w", csv.filename, err)
		}

		fIn.Close()
	}
	f.Close()

	// Replace
	err = os.Rename(tmpFile, csv.filename)
	if err != nil {
		return fmt.Errorf("Could not move temporary file back to %s: %w", csv.filename, err)
	}

	return nil
}

func (csv *CSVDataFile) writeRecordLine(line string) error {
	err := csv.writeHeaders()
	if err != nil {
		return err
	}

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(csv.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Could not open %s: %w", csv.filename, err)
	}
	defer f.Close()

	lineData := []byte(line)
	lineData = append(lineData, '\n')
	if _, err := f.Write(lineData); err != nil {
		return fmt.Errorf("Could not append record line to %s: %w", csv.filename, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("Could not close %s: %w", csv.filename, err)
	}

	return nil
}

func (csv *CSVDataFile) GetHeaderIndex(hdr string) int {
	hdrLen := len(csv.headers)
	for idx, hdrStr := range csv.headers {
		if hdrStr == hdr {
			return idx
		}
	}
	for idx, hdrStr := range csv.headersToAppend {
		if hdrStr == hdr {
			return hdrLen + idx
		}
	}
	csv.headersToAppend = append(csv.headersToAppend, hdr)
	return hdrLen + len(csv.headersToAppend) - 1
}

func (csv *CSVDataFile) WriteRecord(inData map[string]interface{}) error {
	var (
		dataRecord    map[string]interface{} = make(map[string]interface{})
		headerOffsets map[string]int         = make(map[string]int)
		dataLine      []string
		maxOffset     int = 0
	)

	// Shallow-copy data, including timestamp
	for k, v := range inData {
		dataRecord[k] = v
	}
	dataRecord["Timestamp"] = time.Now().Unix()

	// First collect the header indices
	for key, _ := range dataRecord {
		idx := csv.GetHeaderIndex(key)

		if idx > maxOffset {
			maxOffset = idx
		}
		headerOffsets[key] = idx
	}

	// Then create the data line
	dataLine = make([]string, maxOffset+1)
	for key, d := range dataRecord {
		hdrI := headerOffsets[key]

		if v, ok := d.(int); ok {
			dataLine[hdrI] = fmt.Sprintf("%d", v)
		} else if v, ok := d.(uint64); ok {
			dataLine[hdrI] = fmt.Sprintf("%d", v)
		} else if v, ok := d.(int64); ok {
			dataLine[hdrI] = fmt.Sprintf("%d", v)
		} else if v, ok := d.(float64); ok {
			dataLine[hdrI] = fmt.Sprintf("%f", v)
		} else if v, ok := d.(bool); ok {
			if v {
				dataLine[hdrI] = "TRUE"
			} else {
				dataLine[hdrI] = "FALSE"
			}
		} else if v, ok := d.(string); ok {
			if strings.Contains(v, " ") || strings.Contains(v, csv.separator) {
				dataLine[hdrI] = fmt.Sprintf("\"%s\"", v)
			} else {
				dataLine[hdrI] = fmt.Sprintf("%s", v)
			}
		}
	}

	return csv.writeRecordLine(strings.Join(dataLine, csv.separator))
}
