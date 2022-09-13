package sync

type Config struct {
	WorkDir       string `json:"work_dir" required:"true"`
	OBSUtilPath   string `json:"obsutil_path" required:"true"`
	SyncFileShell string `json:"sync_file_shell" required:"true"`

	LFSPath    string `json:"lfs_path" required:"true"`
	RepoPath   string `json:"repo_path" required:"true"`
	Bucket     string `json:"bucket" required:"true"`
	CommitFile string `json:"commit_file" required:"true"`
}
