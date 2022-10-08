package domain

import (
	"errors"
	"regexp"
	"strings"
)

const (
	resourceProject = "project"
	resourceDataset = "dataset"
	resourceModel   = "model"
)

var (
	reName = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

	ResourceTypeProject ResourceType = resourceType(resourceProject)
	ResourceTypeModel   ResourceType = resourceType(resourceModel)
	ResourceTypeDataset ResourceType = resourceType(resourceDataset)
)

// ResourceType
type ResourceType interface {
	ResourceType() string
}

func NewResourceType(v string) (ResourceType, error) {
	if v != resourceProject && v != resourceModel && v != resourceDataset {
		return nil, errors.New("invalid resource type")
	}

	return resourceType(v), nil
}

type resourceType string

func (s resourceType) ResourceType() string {
	return string(s)
}

func ParseResourceType(repoName string) (t ResourceType, err error) {
	if strings.HasPrefix(repoName, resourceProject) {
		t = ResourceTypeProject

	} else if strings.HasPrefix(repoName, resourceModel) {
		t = ResourceTypeModel

	} else if strings.HasPrefix(repoName, resourceDataset) {
		t = ResourceTypeDataset

	} else {
		err = errors.New("unknown repo type")
	}

	return
}

// Account
type Account interface {
	Account() string
}

func NewAccount(v string) (Account, error) {
	if v == "" || strings.ToLower(v) == "root" || !reName.MatchString(v) {
		return nil, errors.New("invalid user name")
	}

	return dpAccount(v), nil
}

type dpAccount string

func (r dpAccount) Account() string {
	return string(r)
}
