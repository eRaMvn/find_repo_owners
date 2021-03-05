package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var owner string
var site string = "github"
var outputFile string
var inputFileSupplied bool
var inputFile string
var token string

// MAX is maximum number of requests at a time
const MAX = 20

// GetAllRepos returns the list of all repos in the organization
func GetAllRepos(ctx *context.Context, client *github.Client) []*github.Repository {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// get all pages of results
	var allRepos []*github.Repository
	for {
		repos, resp, _ := client.Repositories.ListByOrg(*ctx, owner, opt)

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos
}

// GetCodeOwner gets CODEOWNERS file in each repo
func GetCodeOwner(wg *sync.WaitGroup, c chan int, ctx *context.Context, client *github.Client, repo string, knownUsers map[string]bool, resultDict map[string]string) {
	defer wg.Done()
	var contentOptions github.RepositoryContentGetOptions
	resp, err := client.Repositories.DownloadContents(*ctx, owner, repo, "CODEOWNERS", &contentOptions)
	if err != nil {
		// Check a different path if the first path failed
		resp, err = client.Repositories.DownloadContents(*ctx, owner, repo, ".github/CODEOWNERS", &contentOptions)
		if err != nil {
			log.Printf("CODEOWNERS does not exist in '%s' repo", repo)
			url := fmt.Sprintf("https://%s.com/%s/%s", site, owner, repo)
			resultDict[url] = "None"
		} else {
			// Read the response and construct to string
			if b, err := ioutil.ReadAll(resp); err == nil {
				ParseCodeOwners(&b, repo, knownUsers, resultDict)
			}
		}
	} else {
		// Read the response and construct to string
		if b, err := ioutil.ReadAll(resp); err == nil {
			ParseCodeOwners(&b, repo, knownUsers, resultDict)
		}
	}
	<-c
}

// GenerateOwnerString generate owners string and put in resultDict
func GenerateOwnerString(url string, reviewers []string, resultDict map[string]string) {
	allOwners := ""
	for _, reviewer := range reviewers {
		if allOwners == "" {
			allOwners = reviewer
		} else {
			allOwners += fmt.Sprintf("\n%s", reviewer)
		}
	}
	resultDict[url] = allOwners
}

// ParseCodeOwners parses the CODEOWNERS file and return the result to the resultDict
func ParseCodeOwners(b *[]byte, repo string, knownUsers map[string]bool, resultDict map[string]string) {
	linesArray := strings.Split(string(*b), "\n")
	reviewersMap := make(map[string]bool)
	reviewers := []string{}

	url := fmt.Sprintf("https://%s.com/%s/%s", site, owner, repo)

	for _, line := range linesArray {
		// Skip empty line
		if line == "" {
			continue
		}

		lineWords := strings.Split(line, " ")
		// 35 is the # symbol. Skip that
		if lineWords[0][0] == 35 {
			continue
		}

		if inputFileSupplied {
			for _, word := range lineWords {
				if knownUsers[word] {
					resultDict[url] = word
					break
				}
			}
		} else {
			for i, word := range lineWords {
				// Ignore first word
				if i == 0 {
					continue
				}

				if reviewersMap[word] != true {
					reviewers = append(reviewers, word)
					reviewersMap[word] = true
				}
			}
		}
	}

	if inputFileSupplied == false {
		GenerateOwnerString(url, reviewers, resultDict)
	}
}

// ConvertToArray converts hashmap to array of array
func ConvertToArray(resultDict map[string]string) [][]string {
	records := [][]string{
		{"url", "owners"},
	}

	for key, value := range resultDict {
		record := []string{}
		record = append(record, key)
		record = append(record, value)
		records = append(records, record)
	}

	return records
}

// WriteToCSV writes to csv file
func WriteToCSV(records [][]string) {
	fileNameFormatted := fmt.Sprintf("%s.csv", outputFile)
	f, err := os.Create(fileNameFormatted)
	defer f.Close()

	if err != nil {

		log.Fatalln("failed to open file", err)
	}

	w := csv.NewWriter(f)
	err = w.WriteAll(records)

	if err != nil {
		log.Fatal(err)
	}
}

// ReadOwnersFile read file line by line and returns a slice
func ReadOwnersFile() map[string]bool {
	file, err := os.Open(inputFile)

	if err != nil {
		log.Fatalf("failed to open")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	textInput := make(map[string]bool)

	for scanner.Scan() {
		textInput[scanner.Text()] = true
	}

	file.Close()
	return textInput
}

// ExecuteTask runs the task getting owners
func ExecuteTask() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repos := GetAllRepos(&ctx, client)

	// repos := []string{
	// 	// "event-stream",
	// 	// "AV2Foundation",
	// }

	var knownUsers map[string]bool
	// Read from file
	if inputFileSupplied {
		knownUsers = ReadOwnersFile()
	}

	resultDict := make(map[string]string)

	var wg sync.WaitGroup
	c := make(chan int, MAX)

	for _, repo := range repos {
		wg.Add(1)
		c <- 1
		go GetCodeOwner(&wg, c, &ctx, client, *repo.Name, knownUsers, resultDict)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	// Close channel
	close(c)

	WriteToCSV(ConvertToArray(resultDict))
}
