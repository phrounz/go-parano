#!/bin/bash

if [ ! -d vendor/phpmyadmin/sql-parser/ ];then
  which composer >/dev/null
  if [ $? -ne 0 ];then echo "ERROR: missing 'composer', cannot continue" ; exit 1; fi
  composer require phpmyadmin/sql-parser
fi

./go-parano -dir ./examples/ \
  -sql-query-all-in-one \
  -sql-query-func-name 'examplesub.QueryNoAnswer:1,examplesub.Query:2' \
  -sql-query-lint-binary "vendor/phpmyadmin/sql-parser/bin/lint-query" # "./sql-lint-linux"
