#!/usr/bin/env bash

ROOT=$(dirname "${BASH_SOURCE[0]}")
BAD_HEADERS=$((python3 ${ROOT}/verify_boilerplate.py || true) | awk '{ print $7}')

FORMATS="sh go Makefile Dockerfile yaml"

YEAR=`date -u +%Y`

for i in ${FORMATS}
do
	:
	for j in ${BAD_HEADERS}
	do
		:
	        HEADER=$(cat ${ROOT}/boilerplate/boilerplate.${i}.txt | sed "s/YEAR/${YEAR}/")
			value=$(<${j})
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
				echo ${j}
				echo "$text" > ${j}
			fi
	done
done
