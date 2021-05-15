package gitutils

import (
	"fmt"
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

func IsRepo() bool {
	_, err := Exec("rev-parse", "--git-dir")
	return err == nil
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
	return branches([]string{}, func(b string) bool { return true })
}

// RemoteBranches :
func RemoteBranches() []string {
	return branches([]string{"--remote"}, func(branch string) bool { return true })
}

// MergedBranches :
func MergedBranches() []string {
	currentBranch := CurrentBranch()

	return branches([]string{"--merged"}, func(branch string) bool {
		return branch != currentBranch
	})
}

// CurrentBranch :
func CurrentBranch() string {
	out, _ := Exec("branch", "--show-current")
	return out
}

func DefaultBranch() string {
	remote := "origin"
	prefix := remote + "/"

	b := branches([]string{"--remote", "--points-at", prefix + "HEAD"}, func(branch string) bool {
		return strings.HasPrefix(branch, prefix)
	})

	if len(b) > 0 {
		return strings.Replace(b[0], prefix, "", 1)
	}

	return "master"
}

func Fetch() error {
	_, error := Exec("fetch", "--prune", "--tags")
	return error
}

func FetchRemote(remote string) error {
	_, error := Exec("fetch", "--prune", "--tags", remote)
	return error
}

func FetchRemoteInto(remote, branch string) error {
	_, error := Exec("fetch", "--prune", "--tags", "--update-head-ok", remote, fmt.Sprintf("%[1]s:%[1]s", branch))
	return error
}

func ResetHard(ref string) error {
	_, error := Exec("reset", "--hard", ref)
	return error
}

func branches(args []string, filter func(string) bool) []string {
	out, _ := Exec(append([]string{"branch"}, args...)...)

	var branches []string

	for _, branch := range strings.Split(out, "\n") {
		branch = strings.Trim(branch, " *")
		if branch != "" && !strings.Contains(branch, " -> ") && filter(branch) {
			branches = append(branches, branch)
		}
	}

	return branches
}

func IsCurrentBranch(branch string) bool {
	return branch != "" && branch == CurrentBranch()
}

func BranchExists(branch string) bool {
	// TODO use rev-parse? => git rev-parse --abbrev-ref <branch>
	for _, existingBranch := range Branches() {
		if existingBranch == branch {
			return true
		}
	}

	return false
}

func DeleteBranch(branch string) error {
	_, err := Exec("branch", "--delete", branch)
	return err
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

	out, err := Exec("remote", "--verbose")
	re, _ := regexp.Compile(`^([^\s]+)\s+([^\s]+)\s+\((.+)\)$`)

	if err != nil {
		return remotes
	}

	for _, line := range strings.Split(out, "\n") {
		result := re.FindStringSubmatch(line)

		if len(result) == 0 {
			continue
		}

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

func RemoteExists(remoteName string) bool {
	for _, existingRemote := range Remotes() {
		if existingRemote.Name == remoteName {
			return true
		}
	}

	return false
}

func gitCommand(args ...string) *exec.Cmd {

	cmd := exec.Command("git")

	for _, arg := range args {
		cmd.Args = append(cmd.Args, arg)
	}

	return cmd
}
