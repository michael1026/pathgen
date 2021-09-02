package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/michael1026/pathminer/util"
)

func main() {
	var wordlist []string
	urlMap := make(map[string]struct{})
	extensions := []string{""}

	wordlistFile := flag.String("w", "", "Wordlist file")
	extensionsList := flag.String("e", "", "Comma separated list of extensions. Extends FUZZ keyword.")

	flag.Parse()

	if *wordlistFile != "" {
		wordlist, _ = readWordlistIntoFile(*wordlistFile)
	}

	if *extensionsList != "" {
		extensionsSplit := util.SplitByCharAndTrimSpace(*extensionsList, ",")
		for _, ext := range extensionsSplit {
			extensions = util.AppendIfMissing(extensions, ext)
		}
	}

	var urlsToFuzz []string

	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		urlToAdd, _ := url.Parse(s.Text())
		urlToAdd, _ = url.Parse(urlToAdd.Scheme + "://" + urlToAdd.Host + urlToAdd.Path)
		wordlist = wordsFromURL(wordlist, urlToAdd.Path)
		root := urlToAdd.Scheme + "://" + urlToAdd.Host + "/"

		ext := grabExtension(urlToAdd.Path)
		if ext != "" {
			extensions = util.AppendIfMissing(extensions, ext)
		}

		if urlToAdd.Path == "" {
			urlToAdd.Path = "/"
		}

		for root != urlToAdd.String() {
			urlToAdd.Path = path.Dir(urlToAdd.Path)
			if _, ok := urlMap[urlToAdd.String()]; !ok {

				urlsToFuzz = append(urlsToFuzz, urlToAdd.String())
				urlMap[urlToAdd.String()] = struct{}{}
			}
		}

		if _, ok := urlMap[urlToAdd.String()]; !ok {
			urlsToFuzz = append(urlsToFuzz, urlToAdd.String())
			urlMap[urlToAdd.String()] = struct{}{}
		}
	}

	for _, rawUrl := range urlsToFuzz {
		findPaths(rawUrl, &wordlist, &extensions)
	}
}

func findPaths(rawUrl string, wordlist *[]string, extensions *[]string) {
	parsedUrl, _ := url.Parse(rawUrl)
	preStatusParsed, _ := url.Parse(rawUrl)
	preStatusParsed.Path = path.Join(preStatusParsed.Path, util.RandSeq(5))

	fmt.Println(parsedUrl.Path[1:] + "/")
	for _, word := range *wordlist {
		for _, ext := range *extensions {
			newPath := path.Join(parsedUrl.Path, word+ext)
			fmt.Println(newPath[1:])
		}
	}
}

func wordsFromURL(words []string, url string) []string {
	regex := "[A-Za-z]+"

	re := regexp.MustCompile(regex)

	matches := re.FindAllStringSubmatch(url, -1)

	for _, match := range matches {
		words = util.AppendIfMissing(words, match[0])
	}

	return words
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = util.AppendIfMissing(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func readWordlistIntoFile(wordlistPath string) ([]string, error) {
	lines, err := readLines(wordlistPath)
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	return lines, err
}

func grabExtension(path string) string {
	extensions :=
		[]string{".php", ".jsp", ".asp", ".aspx",
			".cgi", ".cfm", ".do", ".go", ".action",
			".axd", ".asmx", ".asx", ".ashx", ".aspx",
			".phtml", ".xhtml"}

	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return ext
		}
	}
	return ""
}
