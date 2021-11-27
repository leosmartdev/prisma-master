#!/bin/bash

ftp -n localhost <<EOF
quote user test
quote pass test
put A029C2900D97591.xml
ren A029C2900D97591.xml A029C2900D97591.txt
EOF

