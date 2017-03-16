# Copyright (c) 2016 heketi authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
# implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from setuptools import setup, find_packages

setup(
    name='heketi',
    version='3.0.0',
    description='Python client library for Heketi',
    license='Apache License (2.0)',
    author='Luis Pabon',
    author_email='lpabon@redhat.com',
    url='https://github.com/heketi/heketi/tree/master/client/api/python',
    packages=find_packages(exclude=['test', 'bin']),
    test_suite='nose.collector',
    install_requires=['pyjwt', 'requests'],
    classifiers=[
        'Development Status :: 5 - Production/Stable'
        'Intended Audience :: Information Technology'
        'Intended Audience :: System Administrators'
        'License :: OSI Approved :: Apache Software License'
        'Operating System :: POSIX :: Linux'
        'Programming Language :: Python'
        'Programming Language :: Python :: 2.7'
    ],
)
