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

// IsIgnored :
func IsIgnored(filePath string) bool {
	_, err := Exec("check-ignore", "--verbose", filePath)

	switch err.(type) {
	case nil:
		return true
	case *exec.ExitError:
		return false
	default:
		panic(err)
	}
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

// Branches :
func Branches() []string {
	out, _ := Exec("branch")

	var branches []string

	for _, branch := range strings.Split(out, "\n") {
		branch = strings.Trim(branch, " *")
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches
}

// RemoteBranches :
func RemoteBranches() []string {
	out, _ := Exec("branch", "--remote")

	var branches []string

	for _, branch := range strings.Split(out, "\n") {
		branch = strings.Trim(branch, " *")
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches
}

// MergedBranches :
func MergedBranches() []string {
	out, _ := Exec("branch", "--merged")

	var branches []string
	var currentBranch = CurrentBranch()

	for _, branch := range strings.Split(out, "\n") {
		branch = strings.Trim(branch, " *")
		if branch != "" && branch != currentBranch {
			branches = append(branches, branch)
		}
	}

	return branches
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
