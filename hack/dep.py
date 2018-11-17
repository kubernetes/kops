#!/usr/bin/env python

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This python script helps sync godeps from the k8s repos into our git submodules
# It generates bash commands where changes are needed
# We can probably also use it for deps when the time comes!

import json
import os
import subprocess
from os.path import join

if not os.environ['GOPATH']:
  raise Exception("Must set GOPATH")

kops_dir = join(os.environ['GOPATH'], 'src', 'k8s.io', 'kops')
k8s_dir = join(os.environ['GOPATH'], 'src', 'k8s.io', 'kubernetes')

with open(join(k8s_dir, 'Godeps', 'Godeps.json')) as data_file:
  godeps = json.load(data_file)

# For debugging, because dep status is unbearably slow
# dep status -json | jq .> dep-status.json
# with open(join(kops_dir, 'dep-status.json')) as data_file:
#   dep_status = json.load(data_file)

process = subprocess.Popen(['dep', 'status', '-json'], stdout=subprocess.PIPE, cwd=kops_dir)
dep_status_stdout, err = process.communicate()
dep_status = json.loads(dep_status_stdout)

#pprint(godeps)

godep_map = {}
for godep in godeps['Deps']:
  #print("%s %s" % (godep['ImportPath'], godep['Rev']))
  godep_map[godep['ImportPath']] = godep['Rev']

dep_status_map = {}
for dep in dep_status:
  #print("%s %s" % (godep['ImportPath'], godep['Rev']))
  dep_status_map[dep['ProjectRoot']] = dep['Revision']


for dep in dep_status_map:
  sha = dep_status_map.get(dep)
  godep_sha = godep_map.get(dep)
  if not godep_sha:
    for k in godep_map:
      if k.startswith(dep):
        godep_sha = godep_map[k]
        break
  if godep_sha:
    if godep_sha != sha:
      print("# update needed: %s %s vs %s" % (dep, godep_sha, sha))
      print("[[override]]")
      print('  name = "%s"' % (dep))
      print('  revision = "%s"' % (godep_sha))
  else:
    print("# UNKNOWN dep %s" % dep)
