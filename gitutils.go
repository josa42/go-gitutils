package gitutils

import (
	"os/exec"
	"regexp"
	"strings"
)

// Remote :
type Remote struct {
	Name  string
	Fetch string
	Push  string
}

// Exec :
func Exec(args ...string) (string, error) {

	cmd := gitCommand(args...)
	outputBytes, err := cmd.Output()

	return strings.Trim(string(outputBytes), " \n"), err
}

// IsDirty :
func IsDirty() bool {

	out, _ := Exec("status", "--porcelain")

	return out != ""
}

// Tags :
func Tags() []string {
	out, _ := Exec("tag", "--list")

	if out != "" {
		// TODO sorting
		return strings.Split(out, "\n")
	}

	return []string{}
}

// AddAll :
func AddAll() {
	Exec("add", "--update")
}

// Commit :
func Commit(message string) error {
	_, error := Exec("commit", "--message", message)

	return error
}

// CommitAll :
func CommitAll(message string) error {
	AddAll()
	return Commit(message)
}

// CommitEmpty :
func CommitEmpty(message string) error {
	_, error := Exec("commit", "--message", message, "--allow-empty")
	return error
}

// Tag :
func Tag(version string) error {
	_, error := Exec("tag", version)

	return error
}

// LastTag :
func LastTag() string {
	out, _ := Exec("describe", "--tags", "--abbrev=0")
	return out
}

// CurrentTag :
func CurrentTag() string {
	out, _ := Exec("describe", "--tags", "--exact")
	return out
}

// CurrentBranch :
func CurrentBranch() string {
	out, _ := Exec("rev-parse", "--abbrev-ref", "HEAD")
	return out
}

// TagExists :
func TagExists(tag string) bool {
	for _, existingTag := range Tags() {
		if existingTag == tag {
			return true
		}
	}

	return false
}

// Push :
func Push() error {
	_, err1 := Exec("push")
	if err1 != nil {
		return err1
	}

	_, err2 := Exec("push", "--tags")
	if err2 != nil {
		return err2
	}

	return nil
}

// Remotes :
func Remotes() map[string]Remote {

	remotes := map[string]Remote{}

	out, _ := Exec("remote", "--verbose")
	re, _ := regexp.Compile(`^([a-z]+)\s+([^\s]+)\s+\((.+)\)$`)

	for _, line := range strings.Split(out, "\n") {
		result := re.FindStringSubmatch(line)

		name := result[1]
		url := result[2]
		urlType := result[3]

		remote := remotes[name]

		if remote.Name == "" {
			remote = Remote{
				Name: name,
			}
		}

		if urlType == "fetch" {
			remote.Fetch = url
		} else if urlType == "push" {
			remote.Push = url
		}

		remotes[name] = remote
	}

	return remotes
}

func gitCommand(args ...string) *exec.Cmd {

	cmd := exec.Command("git")

	for _, arg := range args {
		cmd.Args = append(cmd.Args, arg)
	}

	return cmd
}
