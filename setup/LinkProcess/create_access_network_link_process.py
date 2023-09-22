import json
from dotenv import load_dotenv
import os
import random
import math
import csv

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/LinkProcess"

LINK_SOCKET_ADDRESS_ROOT = "/tmp/mecm2m/link-process/"

# はじめに，元のデータを削除しておく
full_path = json_file_path + "/access_network_link_process.json"
with open(full_path, 'w') as f:
    json.dump({"socket_addresses":[]}, f)

rtt_path = json_file_path + "/rtt.csv"
with open(rtt_path, 'w', newline='') as f:
    writer = csv.writer(f)

psnode_config_file = os.getenv("PROJECT_PATH") + "/setup/GraphDB/config/config_main_psnode.json"
with open(psnode_config_file) as file:
    psnodes = json.load(file)
    psnodes_array = psnodes['psnodes']
    for psnode in psnodes_array:
        psnode_data_property = psnode["psnode"]["data-property"]
        pnode_id = psnode_data_property["PNodeID"]
        # ソケットアドレスからサーバ番号とVNodeIDを抽出
        link_socket_addr = LINK_SOCKET_ADDRESS_ROOT + "access-network_" + pnode_id + ".sock"
        full_path = json_file_path + "/access_network_link_process.json"
        with open(full_path, 'r') as f:
            socket_file_data = json.load(f)
        socket_file_data["socket_addresses"].append(link_socket_addr)
        with open(full_path, 'w') as f:
            json.dump(socket_file_data, f, indent=4)

        # rttファイルに書き込み
        random_number = random.randint(1, 10)
        with open(rtt_path, 'a', newline='') as f:
            writer = csv.writer(f)
            writer.writerow([pnode_id, random_number])