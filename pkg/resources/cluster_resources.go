/*
Copyright 2016 The Kubernetes Authors.

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

package resources

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
)

func DeleteResources(cloud fi.Cloud, resources map[string]*Resource) error {
	depMap := make(map[string][]string)

	done := make(map[string]*Resource)

	var mutex sync.Mutex

	for k, t := range resources {
		for _, block := range t.Blocks {
			depMap[block] = append(depMap[block], k)
		}

		for _, blocked := range t.Blocked {
			depMap[k] = append(depMap[k], blocked)
		}

		if t.Done {
			done[k] = t
		}
	}

	glog.V(2).Infof("Dependencies")
	for k, v := range depMap {
		glog.V(2).Infof("\t%s\t%v", k, v)
	}

	iterationsWithNoProgress := 0
	for {
		// TODO: Some form of default ordering based on types?

		failed := make(map[string]*Resource)

		for {
			phase := make(map[string]*Resource)

			for k, r := range resources {
				if _, d := done[k]; d {
					continue
				}

				if _, d := failed[k]; d {
					// Only attempt each resource once per pass
					continue
				}

				ready := true
				for _, dep := range depMap[k] {
					if _, d := done[dep]; !d {
						glog.V(4).Infof("dependency %q of %q not deleted; skipping", dep, k)
						ready = false
					}
				}
				if !ready {
					continue
				}

				phase[k] = r
			}

			if len(phase) == 0 {
				break
			}

			groups := make(map[string][]*Resource)
			for k, t := range phase {
				groupKey := t.GroupKey
				if groupKey == "" {
					groupKey = "_" + k
				}
				groups[groupKey] = append(groups[groupKey], t)
			}

			var wg sync.WaitGroup
			for _, trackers := range groups {
				wg.Add(1)

				go func(trackers []*Resource) {
					mutex.Lock()
					for _, t := range trackers {
						k := t.Type + ":" + t.ID
						failed[k] = t
					}
					mutex.Unlock()

					defer wg.Done()

					human := trackers[0].Type + ":" + trackers[0].ID

					var err error
					if trackers[0].GroupDeleter != nil {
						err = trackers[0].GroupDeleter(cloud, trackers)
					} else {
						if len(trackers) != 1 {
							glog.Fatalf("found group without groupKey")
						}
						err = trackers[0].Deleter(cloud, trackers[0])
					}
					if err != nil {
						mutex.Lock()
						if IsDependencyViolation(err) {
							fmt.Printf("%s\tstill has dependencies, will retry\n", human)
							glog.V(4).Infof("API call made when had dependency: %s", human)
						} else {
							fmt.Printf("%s\terror deleting resource, will retry: %v\n", human, err)
						}
						for _, t := range trackers {
							k := t.Type + ":" + t.ID
							failed[k] = t
						}
						mutex.Unlock()
					} else {
						mutex.Lock()
						fmt.Printf("%s\tok\n", human)

						iterationsWithNoProgress = 0
						for _, t := range trackers {
							k := t.Type + ":" + t.ID
							delete(failed, k)
							done[k] = t
						}
						mutex.Unlock()
					}
				}(trackers)
			}
			wg.Wait()
		}

		if len(resources) == len(done) {
			return nil
		}

		fmt.Printf("Not all resources deleted; waiting before reattempting deletion\n")
		for k := range resources {
			if _, d := done[k]; d {
				continue
			}

			fmt.Printf("\t%s\n", k)
		}

		iterationsWithNoProgress++
		if iterationsWithNoProgress > 42 {
			return fmt.Errorf("Not making progress deleting resources; giving up")
		}

		time.Sleep(10 * time.Second)
	}
}
