#!/usr/bin/env python3.7
# Copyright 2020 The Cockroach Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Verifies that all source files contain the necessary copyright boilerplate
# snippet.

# This code is based on existing work
# https://github.com/GoogleCloudPlatform/gke-terraform-generator/tree/master/test

from lib.devtools import get_args
from lib.devtools import get_references
from lib.devtools import get_preambles
from lib.devtools import has_validheader

"""
A runnable module to test the presence of boilerplate
text in files within a repo.
"""

def main(args):
    """Identifies and verifies files that should have the desired boilerplate.
    Retrieves the lists of files to be validated and tests each one in turn.
    If all files contain correct boilerplate, this function terminates
    normally. Otherwise it prints the name of each non-conforming file and
    exists with a non-zero status code.
    """
    refs = devtools.get_references(args)
    preambles = devtools.get_preambles(args)
    filenames = devtools.get_files(refs.keys(), args)
    nonconforming_files = []
    for filename in filenames:
        if not devtools.has_valid_header(filename, refs, preambles, REGEXES, args):
            nonconforming_files.append(filename)
    if nonconforming_files:
#        print('%d files have incorrect boilerplate headers:' % len(
#            nonconforming_files))
        for filename in sorted(nonconforming_files):
            print('FAIL: Boilerplate header is wrong for: %s' % os.path.relpath(filename, args.rootdir))
        sys.exit(1)
    else:
        print('All files examined have correct boilerplate.')


if __name__ == "__main__":
    ARGS = get_args()
    main(ARGS)
