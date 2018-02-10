package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	urlFileLocationFlag = flag.String("u", "", "-u list of urls in a new line delimited file. Url must be a valid URL")
	outputFormatFlag    = flag.String("o", "csv", "-o output format. csv or json")
	minTimeFlag         = flag.Int("t", 0, "-t (optional) mintime as epoch will filter results to only show newer than mintime")
	activeFlag          = flag.Bool("a", false, "-a (optional) active flag will try to connect to url on port 443 to see if active")
	print               = fmt.Println
)

const (
	TIME_LAYOUT = "2006-01-02T15:04:05.000Z"
	TIME_SUFFIX = "T00:00:00.000Z"
	// how long to wait before going to next URL
	TIMEOUT_WAIT = 3
)

func main() {

	checkFlags()

	urls, err := getUrls()
	if err != nil {
		log.Fatal(err)
	}

	// get cert sh entries
	var urlEntriesMap = make(map[string][]CertShEntry)
	for _, x := range urls {
		urlEntriesMap[x] = makeRequest(x)
	}

	// filter if mintime flag provided
	if *minTimeFlag != 0 {
		for k, v := range urlEntriesMap {
			var filteredValues []CertShEntry
			for _, x := range v {
				if x.LoggedAtEpoch > *minTimeFlag {
					filteredValues = append(filteredValues, x)
				}
			}
			urlEntriesMap[k] = filteredValues
		}
	}

	// if active, try to connect
	if *activeFlag {
		for k, v := range urlEntriesMap {
			for i, x := range v {
				if tryConnectUrl(x.Hostname) {
					urlEntriesMap[k][i].Active = "Yes"
				} else {
					urlEntriesMap[k][i].Active = "No"
				}
			}
		}
	}

	// create and print results
	var results []Result
	for k, v := range urlEntriesMap {
		entryResults := convertCertShEntryToResult(k, v)
		results = append(results, entryResults...)
		for _, x := range v {
			print("ParentUrl: "+k, "Subdomain: "+x.Hostname, "Active: "+x.Active)
		}
	}

	// save output
	err = saveOutput(results)
	if err != nil {
		log.Fatal(err)
	}

}

func checkFlags() {
	flag.Parse()
	if *urlFileLocationFlag == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *outputFormatFlag != "csv" && *outputFormatFlag != "json" {
		print("Invalid output format. Must be csv or json")
		os.Exit(1)
	}
}

func tryConnectUrl(url string) bool {
	timeout := time.Duration(TIMEOUT_WAIT * time.Second)
	_, err := net.DialTimeout("tcp", url+":443", timeout)
	if err != nil {
		return false
	}
	return true
}

func getUrls() ([]string, error) {

	var retval []string

	file, err := os.Open(*urlFileLocationFlag)
	if err != nil {
		return retval, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		retval = append(retval, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return retval, err
	}

	return retval, nil
}

func prepUrl(rawurl string) string {

	if !strings.HasPrefix(rawurl, "http") {
		rawurl = "http://" + rawurl
	}

	urlStruct, err := url.Parse(rawurl)
	if err != nil {
		log.Fatal(err)
	}

	hostname := urlStruct.Host
	if strings.HasPrefix(hostname, "www.") {
		hostname = hostname[4:]
	}

	template := "https://crt.sh/?q=%25.{{hostname}}"

	hostname = strings.Replace(template, "{{hostname}}", hostname, 1)

	return hostname
}

func makeRequest(rawurl string) []CertShEntry {

	var retval []CertShEntry

	url := prepUrl(rawurl)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var content []string
	var f func(*html.Node)
	// recursively iterate to gather all tables
	f = func(n *html.Node) {
		if n.Parent != nil && n.Parent.Type == html.ElementNode && n.Parent.Data == "td" {
			if strings.Contains(n.Data, "Identity LIKE") {
				return
			} else if n.Data == "a" {
				return
			} else {
				content = append(content, n.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	content = content[2:]

	certShEntry := CertShEntry{}
	for i, x := range content {
		targetIndex := i % 3
		if targetIndex == 0 {
			certShEntry = CertShEntry{}
			certShEntry.Active = "Not tested"
			certShEntry.LoggedAt = x
			certShEntry.LoggedAtEpoch = convertDateStringToEpoch(x)
		} else if targetIndex == 1 {
			certShEntry.NotBefore = x
			certShEntry.NotBeforeEpoch = convertDateStringToEpoch(x)
		} else if targetIndex == 2 {
			certShEntry.Hostname = x
			retval = append(retval, certShEntry)
		}
	}
	return retval
}

func convertDateStringToEpoch(date string) int {

	var retval int

	t, err := time.Parse(TIME_LAYOUT, date+TIME_SUFFIX)
	if err != nil {
		return retval
	}
	return int(t.Unix())
}

func convertCertShEntryToResult(parentUrl string, entries []CertShEntry) []Result {
	var retval []Result

	for _, entry := range entries {
		result := Result{}
		result.Active = entry.Active
		result.LoggedAt = entry.LoggedAt
		result.LoggedAtEpoch = entry.LoggedAtEpoch
		result.NotBefore = entry.NotBefore
		result.NotBeforeEpoch = entry.NotBeforeEpoch

		result.Url = entry.Hostname
		result.ParentUrl = parentUrl
		retval = append(retval, result)
	}

	return retval
}

func saveOutput(results []Result) error {
	if *outputFormatFlag == "json" {
		outBytes, err := json.Marshal(results)
		if err != nil {
			return err
		}
		ioutil.WriteFile("results.json", outBytes, 0644)
		return nil
	} else if *outputFormatFlag == "csv" {
		// header row
		records := [][]string{
			{"ParentUrl", "Url", "Active", "LoggedAt", "NotBefore", "LoggedAtEpoch", "NotBeforeEpoch"},
		}

		for _, x := range results {
			_ = x
			record := []string{
				x.ParentUrl, x.Url, x.Active, x.LoggedAt, x.NotBefore, strconv.Itoa(x.LoggedAtEpoch), strconv.Itoa(x.NotBeforeEpoch)}
			records = append(records, record)
		}

		file, err := os.Create("results.csv")
		if err != nil {
			return err
		}
		defer file.Close()

		w := csv.NewWriter(file)
		for _, record := range records {
			if err := w.Write(record); err != nil {
				log.Fatalln("error writing record to csv:", err)
			}
		}
		defer w.Flush()

		if err := w.Error(); err != nil {
			if err != nil {
				return err
			}
		}
		return nil
	} else {
		return errors.New("Unknown output type")
	}

	return nil
}

type Result struct {
	LoggedAt       string
	NotBefore      string
	ParentUrl      string
	Url            string
	LoggedAtEpoch  int
	NotBeforeEpoch int
	Active         string
}

type CertShEntry struct {
	LoggedAt       string
	NotBefore      string
	Hostname       string
	LoggedAtEpoch  int
	NotBeforeEpoch int
	Active         string
}
