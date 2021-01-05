#!/usr/bin/env bash

# Copyright 2021 The Cockroach Authors
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

ROOT=$(dirname "${BASH_SOURCE[0]}")
# TODO fix this shellcheck error
# shellcheck disable=SC1102
BAD_HEADERS=$((python3 "${ROOT}/verify_boilerplate.py" || true) | awk '{ print $7}')

FORMATS="sh go Makefile Dockerfile yaml"

YEAR=$(date -u +%Y)

for i in ${FORMATS}
do
	:
	for j in ${BAD_HEADERS}
	do
		:
                # TODO fix this shellcheck error
		# shellcheck disable=SC2002
	        HEADER=$(cat "${ROOT}/boilerplate/boilerplate.${i}.txt" | sed "s/YEAR/${YEAR}/")
			value=$(<"${j}")
			if [[ "$j" != *$i ]]
            then
                continue
            fi

			if [[ ${value} == *"# Copyright"* ]]
			then
				echo "Bad header in ${j} ${i}"
			else
				text="$HEADER

$value"
				echo "${j}"
				echo "$text" > "${j}"
			fi
	done
done
