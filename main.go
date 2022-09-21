package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

const (
	repoOwner  string = "github"
	repo       string = "gitignore"
	repoBranch string = "main"
)

var (
	serviceVersion = "dev"

	version         bool
	githubToken     string
	outputDirectory string
	outputFileName  string
	list            bool

	loggerErr = log.New(os.Stderr, "", 0)
	logger    = log.New(os.Stdout, "", 0)
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s [language] [flags]:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&outputFileName, "filename", ".gitignore", "(optional) Output file name.")
	flag.StringVar(&outputDirectory, "directory", ".", "(optional) Output file path.")
	flag.StringVar(&githubToken, "token", os.Getenv("GH_TOKEN"), "(optional) GitHub token to use in case of rate-limited.")
	flag.BoolVar(&list, "list", false, fmt.Sprintf("List available languages on %s/%s@%s.", repoOwner, repo, repoBranch))
	flag.BoolVar(&version, "version", false, "Show version.")
	flag.Parse()

	if version {
		executable, _ := os.Executable()
		logger.Printf("%s version %s", executable, serviceVersion)
		os.Exit(0)
	}

	if len(os.Args) < 2 || os.Args[1] == "" {
		flag.Usage()
		exitWithError("\nplease specify a language")
	}

	ctx := context.TODO()
	client := makeGitHubClient(ctx)

	languages, err := generateAvailableLanguages(ctx, client)
	if err != nil {
		exitWithError("fail to get available languages: %s", err)
	}

	if list {
		printAvailableLanguages(languages)
		os.Exit(0)
	}

	// read language from arg
	language := strings.ToLower(os.Args[1])
	v, ok := languages[language]
	if !ok {
		exitWithError("language %q not found", language)
	}

	content, err := downloadFileContent(ctx, client, v)
	if err != nil {
		exitWithError("fail to download file content: %s", err)
	}

	if err := writeContentToFile(content); err != nil {
		exitWithError("fail to write content to file: %s", err)
	}
}

func exitWithError(format string, v ...any) {
	loggerErr.Printf(format, v...)
	os.Exit(1)
}

// generateAvailableLanguages make a map for all available languages on GitHub
// and their related SHA for further download.
func generateAvailableLanguages(ctx context.Context, client *github.Client) (map[string]string, error) {
	tree, _, err := client.Git.GetTree(ctx, repoOwner, repo, repoBranch, false)
	if err != nil {
		return nil, err
	}

	results := make(map[string]string, len(tree.Entries))
	for _, file := range tree.Entries {
		if file.GetType() != "blob" {
			continue
		}

		if !strings.Contains(file.GetPath(), ".gitignore") {
			continue
		}

		l := strings.TrimSuffix(file.GetPath(), ".gitignore")
		l = strings.ToLower(l)
		results[l] = file.GetSHA()
	}

	return results, nil
}

func downloadFileContent(ctx context.Context, client *github.Client, sha string) ([]byte, error) {
	blob, _, err := client.Git.GetBlob(ctx, repoOwner, repo, sha)
	if err != nil {
		return nil, err
	}

	// content is base64 encoded
	var data []byte
	switch enc := blob.GetEncoding(); enc {
	case "base64":
		data, err = base64.StdEncoding.DecodeString(blob.GetContent())
		if err != nil {
			return nil, err
		}

		return data, nil
	default:
		return nil, fmt.Errorf("encoding %q is not supported", enc)
	}
}

func writeContentToFile(content []byte) error {
	path := filepath.Join(outputDirectory, outputFileName)
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	if _, err := file.Write(content); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	if abs, err := filepath.Abs(path); err == nil {
		logger.Printf("file successfully writed to %s", abs)
	}

	return nil
}

func printAvailableLanguages(data map[string]string) {
	output := "Available languages:\n"
	for k := range data {
		output += fmt.Sprintf("\n%s", k)
	}

	// check if pager is available
	path, err := exec.LookPath(os.Getenv("PAGER"))
	if err != nil {
		logger.Println(output) // raw output if not available
		return
	}

	cmd := exec.Command(path)

	// Feed it with the string you want to display.
	cmd.Stdin = strings.NewReader(output)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logger.Println(output) // raw output if error
	}
}

func makeGitHubClient(ctx context.Context) *github.Client {
	var tc *http.Client
	if githubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
		tc = oauth2.NewClient(ctx, ts)
	}

	return github.NewClient(tc)
}
