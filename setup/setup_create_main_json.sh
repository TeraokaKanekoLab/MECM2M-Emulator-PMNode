# !/bin/bash

# Mainプロセスで使用するconfigファイル (.json) を生成するためのスクリプトファイル
printf "\e[1;31m \n *** CREATE MAIN JSON *** \e[0m"

# ホーム下に用意した.envファイルを読み込む
source $HOME/.env

DIR="/setup/create_main_json"

printf "\e[1;31m \n1. CREATE AREA AND VPOINT JSON \e[0m"
# AreaとVPointのconfigファイルを生成
python3 ${PROJECT_PATH}${DIR}/create_area_json.py

printf "\e[1;31m \n2. CREATE PSINK JSON \e[0m"
# PSinkのconfigファイルを生成
python3 ${PROJECT_PATH}${DIR}/create_psink_json.py

printf "\e[1;31m \n4. CREATE PSNODE AND VSNODE JSON \e[0m"
# PSNodeとVSNODEのconfigファイルを生成
python3 ${PROJECT_PATH}${DIR}/create_psnode_json.py

#printf "\e[1;31m \n5. CREATE PMNODE JSON \e[0m"
# PMNodeのconfigファイルを生成
#python3 ${PROJECT_PATH}${DIR}/create_pmnode_json.py

#printf "\e[1;31m \n6. CREATE PSINK IN PMNODE JSON \e[0m"
# PSink in PMNodeのconfigファイルを生成
#python3 ${PROJECT_PATH}${DIR}/create_psink_in_pmnode_json.py

#printf "\e[1;31m \n7. CREATE PSNODE IN PMNODE JSON \e[0m"
# PSNode in PMNodeのconfigファイルを生成
#python3 ${PROJECT_PATH}${DIR}/create_psnode_in_pmnode_json.py

#printf "\e[1;31m \n8. CREATE EXEC FILE JSON \e[0m"
# 実行形ファイルをまとめたconfigファイルを生成
#python3 ${PROJECT_PATH}${DIR}/create_exec_file_json.py

printf "\n"