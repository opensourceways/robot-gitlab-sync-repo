package obs

/*
type SyncRepoInfo struct {
	WorkDir     string
	RepoUrl     string
	RepoName    string
	StartCommit string
	RepoOBSPath string
}
*/

type OBS interface {
	SaveObject(path, content string) error
	GetObject(path string) ([]byte, error)
	CopyObject(dst, src string) error
	OBSUtilPath() string
}
