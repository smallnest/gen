#!/bin/bash

usage() {
  echo "usage: $0 <DB_TYPE> <APP>"
  echo ""
  echo "    DB_TYPE: sqlite3 | mysql | mssql | postgres"
  echo "    APP: gen_gorm | gen_sqlx | meta"
  exit 0
}

func_gen_gorm() {
  OUT_DIR="./tests/${DB_TYPE}_gorm"
  if [[ -d ${OUT_DIR} ]]; then
    rm -rf "${OUT_DIR}"
  fi

  TABLES_KEY="${DB_TYPE}_tables"
  TABLES=${!TABLES_KEY}

  EXTABLES_KEY="${DB_TYPE}_exclude_tables"
  EXTABLES=${!EXTABLES_KEY}

  if [[ "${TABLES}" == 'all' || "${TABLES}" == '' ]]; then
    echo 'generating gorm code for all tables'
  else
    echo "generating gorm code for tables: ${TABLES}"
    DEFAULT_GEN_OPTIONS="${DEFAULT_GEN_OPTIONS} --table=${TABLES}"
  fi

  if [[ "${EXTABLES}" != '' ]]; then
    echo "excluding tables: ${EXTABLES}"
  fi

  CMD="${GEN_APP} --gorm ${DEFAULT_GEN_OPTIONS} --exclude=\"${EXTABLES}\" --sqltype=\"${DB_TYPE}\" --connstr=\"${DB_CON}\" --database=\"${DB}\" --out=\"${OUT_DIR}\" --module=\"${base_module}/${DB}\""
  eval ${CMD}
}

func_gen_sqlx() {
  OUT_DIR="./tests/${DB_TYPE}_sqlx"
  if [[ -d ./tests/${DB_TYPE} ]]; then
    rm -rf "${OUT_DIR}"
  fi

  TABLES_KEY="${DB_TYPE}_tables"
  TABLES=${!TABLES_KEY}

  EXTABLES_KEY="${DB_TYPE}_exclude_tables"
  EXTABLES=${!EXTABLES_KEY}

  if [[ "${TABLES}" == 'all' || "${TABLES}" == '' ]]; then
    echo 'generating sqlx code for all tables'
  else
    echo "generating sqlx code for tables: ${TABLES}"
    DEFAULT_GEN_OPTIONS="${DEFAULT_GEN_OPTIONS} --table=${TABLES}"
  fi

  if [[ "${EXTABLES}" != '' ]]; then
    echo "excluding tables: ${EXTABLES}"
  fi

  CMD="${GEN_APP} ${DEFAULT_GEN_OPTIONS} --exclude=\"${EXTABLES}\" --sqltype=\"${DB_TYPE}\" --connstr=\"${DB_CON}\" --database=\"${DB}\" --out=\"${OUT_DIR}\" --module=\"${base_module}/${DB}\""
  eval ${CMD}
}

func_meta() {
  DEFAULT_META_OPTIONS=""

  TABLES_KEY="${DB_TYPE}_tables"
  TABLES=${!TABLES_KEY}

  if [[ "${TABLES}" == 'all' || "${TABLES}" == '' ]]; then
    echo 'showing meta for all tables'
  else
    echo "showing meta for tables: ${TABLES}"
    DEFAULT_META_OPTIONS="${DEFAULT_META_OPTIONS} --table=${TABLES}"
  fi

  go run github.com/smallnest/gen/_test/dbmeta \
    --sqltype="${DB_TYPE}" \
    --connstr "${DB_CON}" \
    --database "${DB}" ${DEFAULT_META_OPTIONS}

}

func_test() {
  echo "test"
  echo "${DEFAULT_GEN_OPTIONS}"
  eval print_args ${DEFAULT_GEN_OPTIONS}
}

print_args() {
  echo "--------------------------------"
  echo ""
  args=("$@")
  # get number of elements
  ELEMENTS=${#args[@]}

  # echo each element in array
  # for loop
  for (( i=0;i<${ELEMENTS};i++)); do
    echo "[$i ] ${args[${i}]}"
  done

}
check_vars(){
  if [ "${USE_GEN}" == "" ];
  then
    echo "" >>  ./.env
    echo "USE_GEN=0" >>  ./.env
    . .env
  fi

  if [ "${DEFAULT_GEN_OPTIONS}" == "" ];
  then
    echo "" >>  ./.env
    echo "DEFAULT_GEN_OPTIONS=\"--json
      --db
      --protobuf
      --api=apis
      --dao=daos
      --model=models
      --guregu
      --rest
      --mod
      --server
      --makefile
      --generate-dao
      --generate-proj
      --overwrite
      --copy-templates
      --host=localhost
      --port=8080
      --mapping ./template/mapping.json
      --templateDir=./template
      --file_naming='{{. }}'
      --model_naming='{{FmtFieldName .}}'
      --field_naming='{{FmtFieldName (stringifyFirstChar .) }}'
      --verbose\"" >>  ./.env

      . .env
  fi

}
create_env() {
  cat >./.env <<DELIM
base_module=
USE_GEN=0

postgres_conn=""
postgres_db=
postgres_tables=
postgres_exclude_tables=

mysql_conn=""
mysql_db=
mysql_tables=
mysql_exclude_tables=

sqlite3_conn="./example/sample.db"
sqlite3_db=main
sqlite3_tables=
sqlite3_exclude_tables=

mssql_conn=""
mssql_db=""
mssql_tables=
mssql_exclude_tables=

DEFAULT_GEN_OPTIONS="--json \
      --db \
      --protobuf  \
      --api=apis  \
      --dao=daos  \
      --model=models  \
      --guregu  \
      --rest \
      --mod \
      --server \
      --makefile \
      --generate-dao \
      --generate-proj \
      --overwrite \
      --copy-templates \
      --host=localhost \
      --port=8080 \
      --mapping ./template/mapping.json \
      --templateDir=./template \
      --file_naming='{{. }}' \
      --model_naming='{{FmtFieldName .}}' \
      --field_naming='{{FmtFieldName (stringifyFirstChar .) }}' \
      --verbose"

DELIM

}

if [[ -f ./.env ]]; then
  echo "loaded env file"
  . .env
  check_vars

else
  echo "env file does not exist -= creating template"
  create_env
  exit
fi

DB_TYPE="$1"
APP="$2"

case ${DB_TYPE} in
postgres | mysql | mssql | sqlite3) ;;

*)
  echo "unknown db type: [${DB_TYPE}]"
  usage
  ;;
esac

if [ -z "${base_module}" ]; then
  base_module="example"
fi

DB_CON_NAME="${DB_TYPE}_conn"
DB_NAME="${DB_TYPE}_db"

DB_CON=${!DB_CON_NAME}
DB=${!DB_NAME}

[ -z "${DB_CON}" ] && echo "fill in ${DB_CON_NAME} entry in .env" && exit 0
[ -z "${DB_NAME}" ] && echo "fill in ${DB_NAME} entry in .env" && exit 0

GEN_APP="go run ."
if [ "${USE_GEN}" = "1" ]; then
  GEN_APP=gen
  echo "Using precompiled gen binary"
else
  echo "Using go run - compiling on the fly"
fi

case ${APP} in
gen_gorm | gen_sqlx | meta | test)
  echo "running ${APP}"
  "func_${APP}"
  ;;
*)
  echo "unknown app type: [${APP}]"
  usage
  ;;
esac
