# !/bin/bash

# Mainプロセスで使用するconfigファイル (.json) を生成するためのスクリプトファイル
printf "\e[1;31m \n *** CREATE MAIN JSON *** \e[0m"

# ホーム下に用意した.envファイルを読み込む
source $HOME/.env

DIR="/setup/create_main_json"

printf "\e[1;31m \n1. CREATE VMNODER JSON \e[0m"
# VMNodeRのconfigファイルを生成
python3 ${PROJECT_PATH}${DIR}/create_vmnoder_json.py

printf "\e[1;31m \n2. CREATE PSNODE AND VSNODE JSON \e[0m"
# PSNodeとVSNODEのconfigファイルを生成
python3 ${PROJECT_PATH}${DIR}/create_psnode_json.py

printf "\n"