# !/bin/bash

# 各プロセスの実行形ファイルを生成するためのスクリプト
printf "\e[1;31m \n *** BUILD EXEC FILE *** \e[0m"

# ホーム下に用意した.envファイルを読み込む
source $HOME/.env


printf "\e[1;31m \n1. BUILD MAIN PROCESS \e[0m"
# Mainプロセスの実行系ファイルを生成
go build -o ${PROJECT_PATH}/Main/main ${PROJECT_PATH}/Main/main.go

printf "\e[1;31m \n2. BUILD M2M API PROCESS \e[0m"
# M2M APIプロセスの実行系ファイルを生成
go build -o ${PROJECT_PATH}/M2MAPI/main ${PROJECT_PATH}/M2MAPI/main.go

printf "\e[1;31m \n3. BUILD LOCAL MANAGER PROCESS \e[0m"
# Local Managerプロセスの実行系ファイルを生成
#go build -o ${PROJECT_PATH}/LocalManager/main ${PROJECT_PATH}/LocalManager/main.go

printf "\e[1;31m \n4. BUILD LOCAL AAA PROCESS \e[0m"
# Local AAAプロセスの実行系ファイルを生成
#go build -o ${PROJECT_PATH}/AAA/main ${PROJECT_PATH}/AAA/main.go

printf "\e[1;31m \n5. BUILD LOCAL REPOSITORY PROCESS \e[0m"
# Local Repositoryプロセスの実行系ファイルを生成
#go build -o ${PROJECT_PATH}/LocalRepo/main ${PROJECT_PATH}/LocalRepo/main.go

printf "\e[1;31m \n6. BUILD VSNODE PROCESS \e[0m"
# VSNodeプロセスの実行系ファイルを生成
go build -o ${PROJECT_PATH}/VSNode/main ${PROJECT_PATH}/VSNode/main.go

printf "\e[1;31m \n7. BUILD VMNODE PROCESS \e[0m"
# VMNodeプロセスの実行系ファイルを生成
#go build -o ${PROJECT_PATH}/VMNode/main ${PROJECT_PATH}/VMNode/main.go

printf "\e[1;31m \n8. BUILD PSNODE PROCESS \e[0m"
# PSNodeプロセスの実行系ファイルを生成
go build -o ${PROJECT_PATH}/PSNode/main ${PROJECT_PATH}/PSNode/main.go

printf "\e[1;31m \n9. BUILD ACCESS NETWORK LINK PROCESS \e[0m"
# リンクプロセスの実行系ファイルを生成
go build -o ${PROJECT_PATH}/LinkProcess/main ${PROJECT_PATH}/LinkProcess/main.go

printf "\n"