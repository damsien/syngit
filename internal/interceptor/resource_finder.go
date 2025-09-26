package interceptor

import (
	"fmt"
	"io"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/goccy/go-yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceFinder struct {
	SearchedGVK       schema.GroupVersionResource
	SearchedName      string
	SearchedNamespace string
	Content           string
	paths             []string
}

type ResourceFinderResults struct {
	Found bool
	Paths []string
}

func (rf *ResourceFinder) BuildWorktree(wt *git.Worktree) (ResourceFinderResults, error) {
	rfr := ResourceFinderResults{Found: false, Paths: []string{}}
	rf.paths = []string{}

	err := rf.getPathsContent(wt, wt.Filesystem.Root())
	if err != nil {
		return rfr, err
	}

	if len(rf.paths) > 0 {
		rfr.Found = true
		rfr.Paths = rf.paths
	}

	return rfr, nil
}

func (rf *ResourceFinder) getPathsContent(wt *git.Worktree, basePath string) error {

	files, err := wt.Filesystem.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	var path string
	var currentFileName string

	for _, f := range files {
		currentFileName = f.Name()
		path = basePath + "/" + currentFileName

		if f.IsDir() {
			err = rf.getPathsContent(wt, path)
			if err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(currentFileName, ".yaml") || strings.HasSuffix(currentFileName, ".yml") {

				err = rf.checkInsertResource(wt, path)
				if err != nil {
					return err
				}

			}
		}

	}

	return nil
}

type TypeMeta struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

type ObjectMeta struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type GenericK8sObject struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
}

func (rf *ResourceFinder) checkInsertResource(wt *git.Worktree, path string) error {
	f, err := wt.Filesystem.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open the %s file in the worktree: %w", path, err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read the %s file in the worktree: %w", path, err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close the %s file in the worktree: %w", path, err)
	}

	docs := strings.Split(string(content), "---")

	targetGVK := fmt.Sprintf("%s/%s", rf.SearchedGVK.Group, rf.SearchedGVK.Version)
	if rf.SearchedGVK.Group == "" {
		// Core API group
		targetGVK = rf.SearchedGVK.Version
	}

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var obj GenericK8sObject
		if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
			return fmt.Errorf("failed to unmarshal doc: %w", err)
		}

		if strings.ToLower(obj.Kind) == rf.SearchedGVK.Resource &&
			obj.APIVersion == targetGVK &&
			obj.Name == rf.SearchedName &&
			(rf.SearchedNamespace == "" || obj.Namespace == rf.SearchedNamespace) {

			docs[i] = strings.TrimSpace(string(rf.Content))
			rf.paths = append(rf.paths, path)
			break
		}
	}

	file, err := wt.Filesystem.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create the %s file in the worktree: %w", path, err)
	}

	finalContent := []byte(strings.Join(docs, "\n---\n") + "\n")
	_, err = file.Write(finalContent)
	if err != nil {
		return fmt.Errorf("failed to write the %s file in the worktree: %w", path, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close the %s file in the worktree: %w", path, err)
	}

	return nil
}
