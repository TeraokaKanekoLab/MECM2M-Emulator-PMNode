# !/bin/bash

# register ディレクトリ内のプログラムを順に実行し，GraphDBに初期環境を登録するためのスクリプトファイル
printf "\e[1;31m \n *** REGISTER GRAPHDB *** \e[0m"

# ホーム下に用意した.envファイルを読み込む
source $HOME/.env

DIR="/setup/GraphDB/register"

printf "\e[1;31m \n0. DELETE ALL RECORD \e[0m"
# はじめにGraphDBに登録されているすべてのレコードを削除する
python3 ${PROJECT_PATH}/setup/GraphDB/clear_GraphDB.py

printf "\e[1;31m \n1. REGISTER AREA \e[0m"
# AreaとVPointを登録
python3 ${PROJECT_PATH}${DIR}/1_register_for_neo4j_area_vpoint.py

printf "\e[1;31m \n2. REGISTER PSINK \e[0m"
# PSinkを登録
python3 ${PROJECT_PATH}${DIR}/2_register_for_neo4j_psink.py

printf "\e[1;31m \n4. REGISTER PSNODE AND VSNODE \e[0m"
# PSNodeとVSNODEを登録
python3 ${PROJECT_PATH}${DIR}/3_register_for_neo4j_psnode.py

printf "\n"
