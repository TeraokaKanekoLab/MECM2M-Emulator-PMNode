import json
from dotenv import load_dotenv
import os
import random
import string

# VMNodeR
# ---------------
## Data Property
## * VNodeID
## * Socket Address (モジュールへのアクセス用，本来はIP:Port)
## * Software Module (VNode moduleの実行系ファイル)
## * Description
# ---------------
## Object Property
## * isVirtualizedBy (PNode->VNode)
## * isPhysicalizedBy (VNode->PNode)

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/setup/GraphDB/config/"

vnode_id = str(int(0b1101 << 60))
IP_ADDRESS = os.getenv("IP_ADDRESS")
VMNODER_BASE_PORT = os.getenv("VMNODER_BASE_PORT")
socket_address = IP_ADDRESS + ":" + VMNODER_BASE_PORT
vnode_module = os.getenv("PROJECT_PATH") + "/VMNodeR/main"
description = "Description:VMNodeR"

data = {
    "property-label": "VMNodeR",
    "data-property": {
        "VNodeID": vnode_id, 
        "SocketAddress": socket_address,
        "VNodeModule": vnode_module,
        "Description": description
    }
}

vmnoder_json = json_file_path + "/config_main_vmnoder.json"
with open(vmnoder_json, 'w') as f:
    json.dump(data, f, indent=4)