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

import os
from os import path

gopath=os.environ['GOPATH']

def read_packages_file(package_name):
  packages = []
  with open(path.join(gopath, 'src', package_name, 'hack/.packages')) as packages_file:
    for package in packages_file:
      packages.append(package.replace('\n', ''))
  return packages
