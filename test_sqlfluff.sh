#!/bin/bash

which sqlfluff >/dev/null
if [ $? -ne 0 ];then 
  pip install sqlfluff
  which pip >/dev/null
  if [ $? -ne 0 ];then echo "ERROR: missing 'composer', cannot continue" ; exit 1; fi
fi

./go-parano -dir ./examples/ \
  -sql-query-all-in-one \
  -sql-query-func-name 'examplesub.QueryNoAnswer:1,examplesub.Query:2' \
  -sql-query-lint-binary "sqlfluff lint - --dialect mysql --exclude-rules L006,L008,L009,L013,L039,L011,L031,L036,L003"
# -sql-query-lint-binary "sqlfluff parse --dialect mysql" # <-- may be simpler than using --exclude-rules (keeps only PRS errors AFAIK)
