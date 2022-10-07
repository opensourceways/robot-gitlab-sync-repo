package domain

type RepoSync struct {
	Owner      string
	RepoId     string
	Status     string
	LastCommit string
	version    int
}
