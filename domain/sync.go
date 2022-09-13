package domain

type RepoSync struct {
	Owner      string
	RepoId     string
	LastCommit string
	Status     string
	version    int
}

type Repository interface {
	Find(owner, repoId string) (RepoSync, error)
	Save(*RepoSync) (RepoSync, error)
}
