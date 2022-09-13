package domain

type RepoSync struct {
	Owner      string
	Repo       string
	LastCommit string
	Status     string
	version    int
}

type Repository interface {
	Find(owner, repo string) (RepoSync, error)
	Save(*RepoSync) (RepoSync, error)
}
