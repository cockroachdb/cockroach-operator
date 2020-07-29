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


"""
A runnable module to test the presence of boilerplate
text in files within a repo.
"""

from __future__ import print_function
from subprocess import run, CalledProcessError
import argparse
import glob
import os
import re
import sys


# These directories will be omitted from header checks
SKIPPED_PATHS = [
    'Godeps', 'third_party', '_gopath', '_output',
    '.git', 'vendor', '__init__.py', 'node_modules',
    'bazel-out', 'external', '3rdparty'
]

# A map of regular expressions used in boilerplate validation.
# The date regex is used in validating the date referenced
# is the boilerplate, by ensuring it is an acceptable year.
REGEXES = {
    # beware the Y2100 problem
    "date": re.compile(r'(20\d\d)')
}


def get_args():
    """Parses command line arguments.
    Configures and runs argparse.ArgumentParser to extract command line
    arguments.
    Returns:
        An argparse.Namespace containing the arguments parsed from the
        command line
    """
    parser = argparse.ArgumentParser()

    parser.add_argument("filenames",
                        help="""A list of files to check, all in repo are
                        checked if this is unspecified.""",
                        nargs='*')

    parser.add_argument("-f", "--force-extension",
                        default="",
                        help="""Force an extension to compare against. Useful
                        for files without extensions, such as runnable shell
                        scripts .""")

    parser.add_argument(
        "-r", "--rootdir",
        default=None,
        help="""Root directory of repository. If not specified, the script will
        attempt to draw this value from git.""")

    parser.add_argument("-b", "--boilerplate-dir",
                        default=None,
                        help="""Directory with boilerplate files. Defaults to
                        [root]/hack/boilerplate.""")

    args = parser.parse_args()

    if not args.rootdir:
        ask_git = run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True)
        try:
            ask_git.check_returncode()
        except CalledProcessError:
            print("""No root specfied and directory does not seem to be a git
            repository, or git is not installed.""", file=sys.stderr)
            sys.exit(1)
        args.rootdir = ask_git.stdout.strip()

    if not args.boilerplate_dir:
        args.boilerplate_dir = os.path.join(args.rootdir, "hack/boilerplate")

    return args


def get_references(args):
    """Reads each reference boilerplate file's contents into an array, then
    adds that array to a dictionary keyed by the file extension.

    Returns:
        A dictionary of boilerplate lines, keyed by file extension.
        For example, boilerplate.py.txt would result in the
        k,v pair {".py": py_lines} where py_lines is an array
        containing each line of the file.
    """
    references = {}

    # Find all paths for boilerplate references
    boilerplate_paths = glob.glob(
        os.path.join(args.boilerplate_dir, "boilerplate.*.txt"))

    # Read all boilerplate references into dictionary
    for path in boilerplate_paths:
        with open(path, 'r') as ref_file:
            extension = os.path.basename(path).split(".")[1]
            ref = ref_file.read().splitlines()
            references[extension] = ref

    return references


# Improvement: combine this function with `get_references`
def get_preambles(args):
    """Reads each preamble boilerplate file's contents into an array, then
    adds that array to a dictionary keyed by the file extension.

    Returns:
        A dictionary of boilerplate lines, keyed by file extension.
        For example, boilerplate.py.preamble would result
        in the k,v pair {".py": py_lines} where py_lines is
        an array containing each line of the file
        (ex: "#!/usr/bin/env python3.7")
    """
    preambles = {}

    # Find all paths for boilerplate preambles
    boilerplate_paths = glob.glob(
        os.path.join(args.boilerplate_dir, "boilerplate.*.preamble"))

    # Read all boilerplate preambles into dictionary
    for path in boilerplate_paths:
        with open(path, 'r') as ref_file:
            extension = os.path.basename(path).split(".")[1]
            ref = ref_file.read().splitlines()
            preambles[extension] = ref

    return preambles


def has_valid_header(filename, references, preambles, regexs, args):
    """Test whether a file has the correct boilerplate header.
    Tests each file against the boilerplate stored in refs for that file type
    (based on extension), or by the entire filename (eg Dockerfile, Makefile).
    Some heuristics are applied to remove build tags and shebangs, but little
    variance in header formatting is tolerated.
    Args:
        filename: A string containing the name of the file to test
        references: A map of reference boilerplate text,
            keyed by file extension
        preambles: A map of preamble boilerplate text, keyed by file extension
        regexs: a map of compiled regex objects used in verifying boilerplate
    Returns:
        True if the file has the correct boilerplate header, otherwise returns
        False.
    """
    # Read the entire file.
    with open(filename, 'r') as test_file:
        data = test_file.read()

    # Select the appropriate reference based on the extension,
    #   or if none, the file name.
    basename, extension = get_file_parts(filename)
    if args.force_extension:
        extension = args.force_extension
    elif extension:
        extension = extension
    else:
        extension = basename
    ref = references[extension]
    #print("Verifying boilerplate in file: %s as %s" % (
    #    os.path.relpath(filename, args.rootdir),
    #    extension))

    preamble = preambles.get(extension)
    if preamble:
        preamble = re.escape("\n".join(preamble))
        regflags = re.MULTILINE | re.IGNORECASE
        regex = re.compile(r"^(%s.*\n)\n*" % preamble, regflags)
        (data, _) = regex.subn("", data, 1)

    data = data.splitlines()

    # if our test file is smaller than the reference it surely fails!
    if len(ref) > len(data):
        return False
    # truncate our file to the same number of lines as the reference file
    data = data[:len(ref)]

    # if we don't match the reference at this point, fail
    if ref != data:
        return False

    return True


def get_file_parts(filename):
    """Extracts the basename and extension parts of a filename.
    Identifies the extension as everything after the last period in filename.
    Args:
        filename: string containing the filename
    Returns:
        A tuple of:
            A string containing the basename
            A string containing the extension in lowercase
    """
    extension = os.path.splitext(filename)[1].split(".")[-1].lower()
    basename = os.path.basename(filename)
    return basename, extension


def normalize_files(files, args):
    """Extracts the files that require boilerplate checking from the files
    argument.
    A new list will be built. Each path from the original files argument will
    be added unless it is within one of SKIPPED_DIRS. All relative paths will
    be converted to absolute paths by prepending the root_dir path parsed from
    the command line, or its default value.
    Args:
        files: a list of file path strings
    Returns:
        A modified copy of the files list where any any path in a skipped
        directory is removed, and all paths have been made absolute.
    """
    newfiles = [f for f in files if not any(s in f for s in SKIPPED_PATHS)]

    for idx, pathname in enumerate(newfiles):
        if not os.path.isabs(pathname):
            newfiles[idx] = os.path.join(args.rootdir, pathname)
    return newfiles


def get_files(extensions, args):
    """Generates a list of paths whose boilerplate should be verified.
    If a list of file names has been provided on the command line, it will be
    treated as the initial set to search. Otherwise, all paths within rootdir
    will be discovered and used as the initial set.
    Once the initial set of files is identified, it is normalized via
    normalize_files() and further stripped of any file name whose extension is
    not in extensions.
    Args:
        extensions: a list of file extensions indicating which file types
                    should have their boilerplate verified
    Returns:
        A list of absolute file paths
    """
    files = []
    if args.filenames:
        files = args.filenames
    else:
        for root, dirs, walkfiles in os.walk(args.rootdir):
            # don't visit certain dirs. This is just a performance improvement
            # as we would prune these later in normalize_files(). But doing it
            # cuts down the amount of filesystem walking we do and cuts down
            # the size of the file list
            for dpath in SKIPPED_PATHS:
                if dpath in dirs:
                    dirs.remove(dpath)
            for name in walkfiles:
                pathname = os.path.join(root, name)
                files.append(pathname)
    files = normalize_files(files, args)
    outfiles = []
    for pathname in files:
        basename, extension = get_file_parts(pathname)
        extension_present = extension in extensions or basename in extensions
        if args.force_extension or extension_present:
            outfiles.append(pathname)
    return outfiles
