package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type DocOpDefinition struct {
	ProjectName           string
	SourceDocUrl          string
	SourceDocSectionRegex string
	TargetDocURL          string
	TargetDocSectionRegex string
}

var DocOpDefinitions = []DocOpDefinition{
	{
		ProjectName:           "subfinder",
		SourceDocUrl:          "https://raw.githubusercontent.com/projectdiscovery/subfinder/main/README.md",
		SourceDocSectionRegex: `(?s)##?\s+Usage\n(.*?)(\n##?|\z)`,
		TargetDocURL:          "tools/subfinder/usage.mdx",
		TargetDocSectionRegex: `(?s)##?\s+Subfinder help options\n(.*?)(\n##?|\z)`,
	},
}

func main() {
	updateDoc()
}

func updateDoc() {
	for _, def := range DocOpDefinitions {
		sourceDoc, err := fetchDoc(def.SourceDocUrl)
		if err != nil {
			fmt.Println("Error fetching source doc:", err)
			return
		}
		targetDoc, err := readDoc(def.TargetDocURL)
		if err != nil {
			fmt.Println("Error fetching target doc:", err)
			return
		}

		updatedDoc := replaceSection(sourceDoc, def.SourceDocSectionRegex, targetDoc, def.TargetDocSectionRegex)
		_ = writeDoc(def.TargetDocURL, updatedDoc)
	}
}

func fetchDoc(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch file: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func readDoc(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeDoc(path, data string) error {
	return os.WriteFile(path, []byte(data), 0644)
}

func replaceSection(sourceDoc, sourceDocSectionRegex, targetDoc, targetDocSectionRegex string) string {
	sourceRe := regexp.MustCompile(sourceDocSectionRegex)
	targetRe := regexp.MustCompile(targetDocSectionRegex)

	sourceMatch := sourceRe.FindStringSubmatch(sourceDoc)
	targetMatch := targetRe.FindStringSubmatch(targetDoc)
	if len(sourceMatch) == 0 || len(targetMatch) == 0 {
		return targetDoc
	}

	return strings.Replace(targetDoc, targetMatch[1], sourceMatch[1], 1)
}
