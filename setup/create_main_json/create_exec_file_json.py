from dotenv import load_dotenv
import os
import json
import glob

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/Main/config/json_files"

#EDGE_SERVER_NUM
#CLOUD_SERVER_NUM
#PMNODE_NUM
#PSNODE_NUM
EDGE_SERVER_NUM = int(os.getenv("EDGE_SERVER_NUM"))
PMNODE_NUM = 1
PSNODE_NUM = int(os.getenv("EDGE_SERVER_NUM")) #PSNodeのプロセス数．MEC Server1つに対して1つのPSNodeプロセスを用意する想定．すなわち，MEC Server数

data = {"mec-servers":{"environment":{}}, "pmnodes":{"environment":{}}, "psnodes":{"environment":{}}}

data["mec-servers"]["environment"]["num"] = EDGE_SERVER_NUM
data["pmnodes"]["environment"]["num"] = PMNODE_NUM
data["psnodes"]["environment"]["num"] = PSNODE_NUM

i = 0
data["mec-servers"]["mec-server"] = []
#Serverが使うソケットファイル一覧csvファイルを削除しておく
pattern = os.getenv("PROJECT_PATH") + "/MECServer/Server/*.csv"
matching_files = glob.glob(pattern)
for file_name in matching_files:
    os.remove(file_name)
    #print(f"Remove file: {file_name}")
while i < EDGE_SERVER_NUM:
    server_dict = os.getenv("PROJECT_PATH") + "/MECServer/Server/"
    vpoint_dict = os.getenv("PROJECT_PATH") + "/MECServer/VPoint/"
    vsnode_dict = os.getenv("PROJECT_PATH") + "/MECServer/VSNode/"
    vmnode_dict = os.getenv("PROJECT_PATH") + "/MECServer/VMNode/"
    main_file = "main"
    server_socket_file = server_dict + "mec_server_" + str(i+1) + ".csv"
    with open(server_socket_file, "w") as file:
        #m2mApi, localMgr, pnodeMgr, aaa, localRepo, graphDB, sensingDB
        m2mApi_sock_addr    =    "/tmp/mecm2m/svr_" + str(i+1) + "_m2mapi.sock"
        localMgr_sock_addr  =    "/tmp/mecm2m/svr_" + str(i+1) + "_localmgr.sock"
        pnodeMgr_sock_addr  =    "/tmp/mecm2m/svr_" + str(i+1) + "_pnodemgr.sock"
        aaa_sock_addr       =    "/tmp/mecm2m/svr_" + str(i+1) + "_aaa.sock"
        localRepo_sock_addr =    "/tmp/mecm2m/svr_" + str(i+1) + "_localrepo.sock"
        graphDB_sock_addr   =    "/tmp/mecm2m/svr_" + str(i+1) + "_graphdb.sock"
        sensingDB_sock_addr =    "/tmp/mecm2m/svr_" + str(i+1) + "_sensingdb.sock"
        file.write(f"m2mApi,{m2mApi_sock_addr}\n")
        file.write(f"localMgr,{localMgr_sock_addr}\n")
        file.write(f"pnodeMgr,{pnodeMgr_sock_addr}\n")
        file.write(f"aaa,{aaa_sock_addr}\n")
        file.write(f"localRepo,{localRepo_sock_addr}\n")
        file.write(f"graphDB,{graphDB_sock_addr}\n")
        file.write(f"sensingDB,{sensingDB_sock_addr}\n")
    mec_server_dict = {
        "Server": server_dict + main_file,
        "VPoint": vpoint_dict + main_file,
        "VSNode": vsnode_dict + main_file,
        "vmnode": vmnode_dict + main_file
    }
    data["mec-servers"]["mec-server"].append(mec_server_dict)
    i += 1

i = 0
data["pmnodes"]["pmnode"] = []
while i < PMNODE_NUM:
    mserver_dict = os.getenv("PROJECT_PATH") + "/PMNode/MServer/"
    vpoint_dict = os.getenv("PROJECT_PATH") + "/PMNode/VPoint/"
    vsnode_dict = os.getenv("PROJECT_PATH") + "/PMNode/VSNode/"
    psnode_dict = os.getenv("PROJECT_PATH") + "/PMNode/PSNode/"
    main_file = "main"
    #MServerで使う，ソケットファイル名をまとめたファイルをここで作る
    pmnode_dict = {
        "MServer": mserver_dict + main_file,
        "VPoint": vpoint_dict + main_file,
        "VSNode": vsnode_dict + main_file,
        "PSNode": psnode_dict + main_file,
        "MServer-config-file": mserver_dict + "config.json"     #MServer用のconfigファイル
    }
    data["pmnodes"]["pmnode"].append(pmnode_dict)
    i += 1

i = 0
data["psnodes"]["psnode"] = []
while i < PSNODE_NUM:
    psnode_dict = os.getenv("PROJECT_PATH") + "/PSNode/"
    psnode_dict = {
        "PSNode": psnode_dict + "main",
        "config-file": psnode_dict + "config.json"              #PSNode用のconfigファイル
    }
    data["psnodes"]["psnode"].append(psnode_dict)
    i += 1

exec_file_json = json_file_path + "/config_main_exec_file.json"
with open(exec_file_json, 'w') as f:
    json.dump(data, f, indent=4)
