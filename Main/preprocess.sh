# !/bin/bash

# Mainプロセスを実行する前にデータベースなどを事前に準備するためのスクリプト
printf "\e[1;31m \n *** PREPROCESSING *** \e[0m"

# ホーム下に用意した.envファイルを読み込む
source $HOME/.env

DIR="/setup/SensingDB"

printf "\e[1;31m \n1. CREATE CONFIG FILE \e[0m"
# 各種インスタンスのconfigファイルを生成
bash ${PROJECT_PATH}/setup/setup_create_main_json.sh

printf "\e[1;31m \n2. REGISTER FOR GRAPHDB \e[0m"
# GraphDB にデータを保存
bash ${PROJECT_PATH}/setup/GraphDB/register_GraphDB.sh

printf "\e[1;31m \n3. CREATE LINK PROCSS \e[0m"
# リンクプロセスの作成
python3 ${PROJECT_PATH}/setup/LinkProcess/create_access_network_link_process.py

printf "\n"