#!/bin/bash

if [ ! -d vendor/phpmyadmin/sql-parser/ ];then
  which composer >/dev/null
  if [ $? -ne 0 ];then echo "ERROR: missing 'composer', cannot continue" ; exit 1; fi
  composer require phpmyadmin/sql-parser
fi

./go-parano.out -dir ./examples/ \
  -sql-query-func-name examplesub.Query,examplesub.QueryNoAnswer \
  -sql-query-lint-binary "vendor/phpmyadmin/sql-parser/bin/lint-query" # "./sql-lint-linux"
