/* Copyright 2019 The Bazel Authors. All rights reserved.

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

package repo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/bazelbuild/bazel-gazelle/label"
)

type goDepLockFile struct {
	ImportPath   string
	GoVersion    string
	GodepVersion string
	Packages     []string
	Deps         []goDepProject
}

type goDepProject struct {
	ImportPath string
	Rev        string
}

func importRepoRulesGoDep(filename string, cache *RemoteCache) ([]Repo, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	file := goDepLockFile{}

	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	roots := make([]string, len(file.Deps))
	errs := make([]error, len(file.Deps))

	wg.Add(len(file.Deps))
	for i, p := range file.Deps {
		go func(i int, p goDepProject) {
			defer wg.Done()
			rootRepo, _, err := cache.Root(p.ImportPath)
			if err != nil {
				errs[i] = err
			} else {
				roots[i] = rootRepo
			}
		}(i, p)
	}
	wg.Wait()

	var repos []Repo
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	repoToRev := make(map[string]string)

	for i, p := range file.Deps {
		repoRoot := roots[i]
		if rev, ok := repoToRev[repoRoot]; !ok {
			repos = append(repos, Repo{
				Name:     label.ImportPathToBazelRepoName(repoRoot),
				GoPrefix: repoRoot,
				Commit:   p.Rev,
			})
			repoToRev[repoRoot] = p.Rev
		} else {
			if p.Rev != rev {
				return nil, fmt.Errorf("Repo %s imported at multiple revisions: %s, %s", repoRoot, p.Rev, rev)
			}
		}
	}
	return repos, nil
}
