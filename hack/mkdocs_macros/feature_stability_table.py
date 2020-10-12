#!/usr/bin/env python

# Copyright 2020 The Kubernetes Authors.
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


def define_env(env):
    """Hook function"""

    @env.macro
    def kops_feature_table(**kwargs):
        """
        Generate a markdown table which will be rendered when called, along with the supported passed keyword args.
        :param kwargs:
                       kops_added_ff => Kops version in which this feature was added as a feature flag
                       kops_added_default => Kops version in which this feature was introduced as stable
                       k8s_min => Minimum k8s version which supports this feature
        :return: rendered markdown table
        """

        # this dict object maps the kwarg to its description, which will be used in the final table
        supported_args = {
            'kops_added_ff': 'Alpha (Feature Flag)',
            'kops_added_default': 'Default',
            'k8s_min': 'Minimum K8s Version'
        }

        # Create the initial strings to which we'll concatenate the relevant columns
        title = '|'
        separators = '|'
        values = '|'

        # Iterate over provided supported kwargs and match them with the provided values.
        for arg, header in supported_args.items():
            if arg not in kwargs.keys():
                continue
            if arg == 'kops_added_default' and 'kops_added_ff' not in kwargs.keys():
                title += ' Introduced |'
            else:
                title += f' {header} |'
            separators += ' :-: |'
            if arg == 'k8s_min':
                values += f' K8s {kwargs[arg]} |'
            else:
                values += f' Kops {kwargs[arg]} |'

        # Create a list object containing all the table rows,
        # Then return a string object which contains every list item in a new line.
        table = [
            title,
            separators,
            values
        ]
        return '\n'.join(table)


def main():
    pass


if __name__ == "__main__":
    main()
