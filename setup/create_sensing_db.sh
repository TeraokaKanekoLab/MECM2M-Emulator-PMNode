# !/bin/bash

# SensingDB ディレクトリ内のプログラムを実行し，MySQLに新規データベースを作成するためのスクリプトファイル
printf "\e[1;31m \n *** CREATE SENSINGDB *** \e[0m"

# ホーム下に用意した.envファイルを読み込む
source $HOME/.env

DIR="/setup/SensingDB"

printf "\e[1;31m \n0. DELETE ALL RECORD \e[0m"
# はじめにSensingDBに登録されているすべてのレコードとテーブルを削除する
python3 ${PROJECT_PATH}${DIR}/clear_SensingDB.py

printf "\e[1;31m \n1. CREATE DB TABLE INDEX \e[0m"
# データベース，テーブル，インデックスを作成
python3 ${PROJECT_PATH}${DIR}/create_db_table_index.py

printf "\n"