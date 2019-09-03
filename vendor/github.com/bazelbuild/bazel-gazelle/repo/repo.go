/* Copyright 2017 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package repo provides functionality for managing Go repository rules.
//
// UNSTABLE: The exported APIs in this package may change. In the future,
// language extensions should implement an interface for repository
// rule management. The update-repos command will call interface methods,
// and most if this package's functionality will move to language/go.
// Moving this package to an internal directory would break existing
// extensions, since RemoteCache is referenced through the resolve.Resolver
// interface, which extensions are required to implement.
package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/merger"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// Repo describes an external repository rule declared in a Bazel
// WORKSPACE file or macro file.
type Repo struct {
	// Name is the value of the "name" attribute of the repository rule.
	Name string

	// GoPrefix is the portion of the Go import path for the root of this
	// repository. Usually the same as Remote.
	GoPrefix string

	// Commit is the revision at which a repository is checked out (for example,
	// a Git commit id).
	Commit string

	// Tag is the name of the version at which a repository is checked out.
	Tag string

	// Remote is the URL the repository can be cloned or checked out from.
	Remote string

	// VCS is the version control system used to check out the repository.
	// May also be "http" for HTTP archives.
	VCS string

	// Version is the semantic version of the module to download. Exactly one
	// of Version, Commit, and Tag must be set.
	Version string

	// Sum is the hash of the module to be verified after download.
	Sum string

	// Replace is the Go import path of the module configured by the replace
	// directive in go.mod.
	Replace string
}

type byName []Repo

func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type byRuleName []*rule.Rule

func (s byRuleName) Len() int           { return len(s) }
func (s byRuleName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s byRuleName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type lockFileFormat int

const (
	unknownFormat lockFileFormat = iota
	depFormat
	moduleFormat
	godepFormat
)

var lockFileParsers = map[lockFileFormat]func(string, *RemoteCache) ([]Repo, error){
	depFormat:    importRepoRulesDep,
	moduleFormat: importRepoRulesModules,
	godepFormat:  importRepoRulesGoDep,
}

// ImportRepoRules reads the lock file of a vendoring tool and returns
// a list of equivalent repository rules that can be merged into a WORKSPACE
// file. The format of the file is inferred from its basename.
func ImportRepoRules(filename string, repoCache *RemoteCache) ([]*rule.Rule, error) {
	format := getLockFileFormat(filename)
	if format == unknownFormat {
		return nil, fmt.Errorf(`%s: unrecognized lock file format. Expected "Gopkg.lock", "go.mod", or "Godeps.json"`, filename)
	}
	parser := lockFileParsers[format]
	repos, err := parser(filename, repoCache)
	if err != nil {
		return nil, fmt.Errorf("error parsing %q: %v", filename, err)
	}
	sort.Stable(byName(repos))

	rules := make([]*rule.Rule, 0, len(repos))
	for _, repo := range repos {
		rules = append(rules, GenerateRule(repo))
	}
	return rules, nil
}

// MergeRules merges a list of generated repo rules with the already defined repo rules,
// and then updates each rule's underlying file. If the generated rule matches an existing
// one, then it inherits the file where the existing rule was defined. If the rule is new then
// its file is set as the destFile parameter. If pruneRules is set, then this function will prune
// any existing rules that no longer have an equivalent repo defined in the Gopkg.lock/go.mod file.
// A list of the updated files is returned.
func MergeRules(genRules []*rule.Rule, existingRules map[*rule.File][]string, destFile *rule.File, kinds map[string]rule.KindInfo, pruneRules bool) []*rule.File {
	sort.Stable(byRuleName(genRules))

	ruleMap := make(map[string]bool)
	if pruneRules {
		for _, r := range genRules {
			ruleMap[r.Name()] = true
		}
	}

	repoMap := make(map[string]*rule.File)
	emptyRules := make([]*rule.Rule, 0)
	for file, repoNames := range existingRules {
		// Avoid writing to the same file by matching destFile with its definition in existingRules
		if file.Path == destFile.Path && file.MacroName() != "" && file.MacroName() == destFile.MacroName() {
			file = destFile
		}
		for _, name := range repoNames {
			if pruneRules && !ruleMap[name] {
				emptyRules = append(emptyRules, rule.NewRule("go_repository", name))
			}
			repoMap[name] = file
		}
	}

	rulesByFile := make(map[*rule.File][]*rule.Rule)
	for _, rule := range genRules {
		dest := destFile
		if file, ok := repoMap[rule.Name()]; ok {
			dest = file
		}
		rulesByFile[dest] = append(rulesByFile[dest], rule)
	}
	emptyRulesByFile := make(map[*rule.File][]*rule.Rule)
	for _, rule := range emptyRules {
		if file, ok := repoMap[rule.Name()]; ok {
			emptyRulesByFile[file] = append(emptyRulesByFile[file], rule)
		}
	}

	updatedFiles := make(map[string]*rule.File)
	for f, rules := range rulesByFile {
		merger.MergeFile(f, emptyRulesByFile[f], rules, merger.PreResolve, kinds)
		delete(emptyRulesByFile, f)
		f.Sync()
		if uf, ok := updatedFiles[f.Path]; ok {
			uf.SyncMacroFile(f)
		} else {
			updatedFiles[f.Path] = f
		}
	}
	// Merge the remaining files that have empty rules, but no genRules
	for f, rules := range emptyRulesByFile {
		merger.MergeFile(f, rules, nil, merger.PreResolve, kinds)
		f.Sync()
		if uf, ok := updatedFiles[f.Path]; ok {
			uf.SyncMacroFile(f)
		} else {
			updatedFiles[f.Path] = f
		}
	}

	files := make([]*rule.File, 0, len(updatedFiles))
	for _, f := range updatedFiles {
		files = append(files, f)
	}
	return files
}

func getLockFileFormat(filename string) lockFileFormat {
	switch filepath.Base(filename) {
	case "Gopkg.lock":
		return depFormat
	case "go.mod":
		return moduleFormat
	case "Godeps.json":
		return godepFormat
	default:
		return unknownFormat
	}
}

// GenerateRule returns a repository rule for the given repository that can
// be written in a WORKSPACE file.
func GenerateRule(repo Repo) *rule.Rule {
	r := rule.NewRule("go_repository", repo.Name)
	if repo.Commit != "" {
		r.SetAttr("commit", repo.Commit)
	}
	if repo.Tag != "" {
		r.SetAttr("tag", repo.Tag)
	}
	r.SetAttr("importpath", repo.GoPrefix)
	if repo.Remote != "" {
		r.SetAttr("remote", repo.Remote)
	}
	if repo.VCS != "" {
		r.SetAttr("vcs", repo.VCS)
	}
	if repo.Version != "" {
		r.SetAttr("version", repo.Version)
	}
	if repo.Sum != "" {
		r.SetAttr("sum", repo.Sum)
	}
	if repo.Replace != "" {
		r.SetAttr("replace", repo.Replace)
	}
	return r
}

// FindExternalRepo attempts to locate the directory where Bazel has fetched
// the external repository with the given name. An error is returned if the
// repository directory cannot be located.
func FindExternalRepo(repoRoot, name string) (string, error) {
	// See https://docs.bazel.build/versions/master/output_directories.html
	// for documentation on Bazel directory layout.
	// We expect the bazel-out symlink in the workspace root directory to point to
	// <output-base>/execroot/<workspace-name>/bazel-out
	// We expect the external repository to be checked out at
	// <output-base>/external/<name>
	// Note that users can change the prefix for most of the Bazel symlinks with
	// --symlink_prefix, but this does not include bazel-out.
	externalPath := strings.Join([]string{repoRoot, "bazel-out", "..", "..", "..", "external", name}, string(os.PathSeparator))
	cleanPath, err := filepath.EvalSymlinks(externalPath)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%s: not a directory", externalPath)
	}
	return cleanPath, nil
}

// ListRepositories extracts metadata about repositories declared in a
// file.
func ListRepositories(workspace *rule.File) (repos []Repo, repoNamesByFile map[*rule.File][]string, err error) {
	repoNamesByFile = make(map[*rule.File][]string)
	repos, repoNamesByFile[workspace] = getRepos(workspace.Rules)
	for _, d := range workspace.Directives {
		switch d.Key {
		case "repository_macro":
			f, defName, err := parseRepositoryMacroDirective(d.Value)
			if err != nil {
				return nil, nil, err
			}
			f = filepath.Join(filepath.Dir(workspace.Path), filepath.Clean(f))
			macroFile, err := rule.LoadMacroFile(f, "", defName)
			if err != nil {
				return nil, nil, err
			}
			currRepos, names := getRepos(macroFile.Rules)
			repoNamesByFile[macroFile] = names
			repos = append(repos, currRepos...)
		}
	}

	return repos, repoNamesByFile, nil
}

func parseRepositoryMacroDirective(directive string) (string, string, error) {
	vals := strings.Split(directive, "%")
	if len(vals) != 2 {
		return "", "", fmt.Errorf("Failure parsing repository_macro: %s, expected format is macroFile%%defName", directive)
	}
	f := vals[0]
	if strings.HasPrefix(f, "..") {
		return "", "", fmt.Errorf("Failure parsing repository_macro: %s, macro file path %s should not start with \"..\"", directive, f)
	}
	return f, vals[1], nil
}

func getRepos(rules []*rule.Rule) (repos []Repo, names []string) {
	for _, r := range rules {
		name := r.Name()
		if name == "" {
			continue
		}
		var repo Repo
		switch r.Kind() {
		case "go_repository":
			// TODO(jayconrod): extract other fields needed by go_repository.
			// Currently, we don't use the result of this function to produce new
			// go_repository rules, so it doesn't matter.
			goPrefix := r.AttrString("importpath")
			version := r.AttrString("version")
			sum := r.AttrString("sum")
			replace := r.AttrString("replace")
			revision := r.AttrString("commit")
			tag := r.AttrString("tag")
			remote := r.AttrString("remote")
			vcs := r.AttrString("vcs")
			if goPrefix == "" {
				continue
			}
			repo = Repo{
				Name:     name,
				GoPrefix: goPrefix,
				Version:  version,
				Sum:      sum,
				Replace:  replace,
				Commit:   revision,
				Tag:      tag,
				Remote:   remote,
				VCS:      vcs,
			}

			// TODO(jayconrod): infer from {new_,}git_repository, {new_,}http_archive,
			// local_repository.

		default:
			continue
		}
		repos = append(repos, repo)
		names = append(names, repo.Name)
	}
	return repos, names
}
