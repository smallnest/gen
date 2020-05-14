#!/bin/bash


usage() {
  echo "usage: $0 <DB_TYPE> <APP>"
  echo ""
  echo "    DB_TYPE: sqlite3 | mysql | mssql | postgres"
  echo "    APP: gen | meta"
  exit 0
}

func_gen() {
  if [[ -d ./tests/${DB_TYPE} ]];
  then
    rm -rf "./tests/${DB_TYPE}"
  fi

  DEFAULT_GEN_OPTIONS="--json
      --api=apis
      --dao=daos
      --model=models
      --gorm
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
      --db
      --protobuf
      --templateDir=./template
      --verbose"



  go run . \
    --sqltype="${DB_TYPE}" \
    --connstr="${DB_CON}" \
    --database="${DB}" \
    --out="./tests/${DB_TYPE}" \
    --module="github.com/alexj212A/${DB}" \
    ${DEFAULT_GEN_OPTIONS}
}

func_meta() {
  DEFAULT_META_OPTIONS=""

  go run github.com/smallnest/gen/_test/dbmeta \
    --sqltype="${DB_TYPE}" \
    --connstr "${DB_CON}" \
    --database "${DB}" \
    ${DEFAULT_META_OPTIONS}
}




create_env(){
cat > ./.env <<DELIM

postgres_conn=""
postgres_db=

mysql_conn=""
mysql_db=

sqlite3_conn="./example/sample.db"
sqlite3_db=main

mssql_conn=""
mssql_db=""

DELIM

}



if [[ -f ./.env ]];
then
  echo "loaded env file"
  . .env
else
  echo "env file does not exist -= creating template"
  create_env
  exit
fi



DB_TYPE="$1"
APP="$2"

case ${DB_TYPE} in
postgres | mysql | mssql | sqlite3)
  ;;
*)
  echo "unknown db type: [${DB_TYPE}]"
  usage
  ;;
esac


DB_CON_NAME="${DB_TYPE}_conn"
DB_NAME="${DB_TYPE}_db"

DB_CON=${!DB_CON_NAME}
DB=${!DB_NAME}


[ -z "${DB_CON}" ] && echo "fill in ${DB_CON_NAME} entry in .env" && exit 0
[ -z "${DB_NAME}" ] && echo "fill in ${DB_NAME} entry in .env" && exit 0


case ${APP} in
gen | meta)
  echo "running ${APP}"
  "func_${APP}"
  ;;
*)
  echo "unknown app type: [${APP}]"
  usage
  ;;
esac


