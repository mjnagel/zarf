package git

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"
)

type GitCredential struct {
	path string
	auth http.BasicAuth
}

func transformURLtoRepoName(url string) string {
	replaceRegex := regexp.MustCompile(`^[^@]+@|^https:\/\/|[^\w\-\.]+`)
	return "mirror" + replaceRegex.ReplaceAllString(url, "__")
}

func transformURL(baseUrl string, url string) string {
	replaced := transformURLtoRepoName(url)
	return baseUrl + "/syncuser/" + replaced
}

func credentialFilePath() string {
	homePath, _ := os.UserHomeDir()
	return homePath + "/.git-credentials"
}

func credentialParser() []GitCredential {
	credentialsPath := credentialFilePath()
	credentials := []GitCredential{}

	credentialsFile, err := os.Open(credentialsPath)
	if err != nil {
		logrus.Info(credentialsPath + " file not found")
	}
	defer credentialsFile.Close()

	scanner := bufio.NewScanner(credentialsFile)
	for scanner.Scan() {
		gitUrl, err := url.Parse(scanner.Text())
		password, _ := gitUrl.User.Password()
		if err != nil {
			continue
		}
		credential := GitCredential{
			path: gitUrl.Host,
			auth: http.BasicAuth{
				Username: gitUrl.User.Username(),
				Password: password,
			},
		}
		credentials = append(credentials, credential)
	}

	if err := scanner.Err(); err != nil {
		logrus.Warn("Error parsing git credentials file")
	}

	return credentials
}

func findAuthForHost(baseUrl string) GitCredential {
	// Read the ~/.git-credentials file
	gitCreds := credentialParser()

	// Will be nil unless a match is found
	var matchedCred GitCredential

	// Look for a match for the given host path in the creds file
	for _, gitCred := range gitCreds {
		hasPath := strings.Contains(baseUrl, gitCred.path)
		if hasPath {
			matchedCred = gitCred
			break
		}
	}

	return matchedCred
}

func CredentialsGenerator(host string, username string, password string) {
	credentialsPath := credentialFilePath()

	// Prevent duplicates by purging the git creds file~
	_ = os.Remove(credentialsPath)

	credentialsFile, err := os.OpenFile(credentialsPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logrus.Fatal("Unable to access the git credentials file")
	}
	defer credentialsFile.Close()

	credentialsText := fmt.Sprintf("https://%s:%s@%s\n", username, password, host)

	// Write the entry to the file
	_, err = credentialsFile.WriteString(credentialsText)
	if err != nil {
		logrus.Fatal("Unable to update the git credentials file")
	}

	// Save the change
	err = credentialsFile.Sync()
	if err != nil {
		logrus.Fatal("Unable to update the git credentials file")
	}
}