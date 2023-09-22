import json
from dotenv import load_dotenv
import os
import math
import random
import string
import glob

# PSNode
# ---------------
## Data Property
## * PNodeID
## * PNode Type (デバイスとクラスの対応づけ)
## * VNode Module (VNode moduleの実行系ファイル)
## * Socket Address (デバイスへのアクセス用，本来はIP:Port)
## * Lat Lon
## * Capability
## * Credential
## * Session Key
## * Description 
# ---------------
## Object Property
## * aggregates (PSink->PNode)
## * isConnectedTo (PNode->PSink)
## * contains (PArea->PNode)
## * isInstalledIn (PNode->PArea)

# VSNode
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

# session key 生成
def generate_random_string(length):
    # 文字列に使用する文字を定義
    letters = string.ascii_letters + string.digits

    result_str = ''.join(random.choice(letters) for i in range(length))

    return result_str

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/setup/GraphDB/config/"

# VSNODE_BASE_PORT
# PSINK_NUM_PER_AREA
# MIN_LAT, MAX_LAT, MIN_LON, MAX_LON
# AREA_WIDTH
# IP_ADDRESS
VSNODE_BASE_PORT = int(os.getenv("VSNODE_BASE_PORT"))
PSINK_NUM_PER_AREA = float(os.getenv("PSINK_NUM_PER_AREA"))
MIN_LAT = float(os.getenv("MIN_LAT"))
MAX_LAT = float(os.getenv("MAX_LAT"))
MIN_LON = float(os.getenv("MIN_LON"))
MAX_LON = float(os.getenv("MAX_LON"))
AREA_WIDTH = float(os.getenv("AREA_WIDTH"))
IP_ADDRESS = os.getenv("IP_ADDRESS")

lineStep = AREA_WIDTH
forint = 1000

data = {"psnodes":[]}

# 始点となるArea
swLat = MIN_LAT
neLat = swLat + lineStep

# label情報
label_lat = 0
label_lon = 0

# ID用のindex
id_index = 0

# psink用のindex
psink_index = 0

# PNTypeをあらかじめ用意
pn_types = ["Temp_Sensor", "Humid_Sensor", "Anemometer"]
capabilities = {"Temp_Sensor":"MaxTemp", "Humid_Sensor":"MaxHumid", "Anemometer":"MaxWind"}
vsnode_psnode_relation = {}
for i in range(len(pn_types)):
    vsnode_psnode_relation[pn_types[i]] = []

# PSNode/initial_environment.json の初期化
psnode_dir_path = os.getenv("PROJECT_PATH") + "/PSNode/"
initial_environment_file = psnode_dir_path + "initial_environment.json"
port_array = {
    "ports": []
}
with open(initial_environment_file, 'w') as file:
    json.dump(port_array, file, indent=4)

# VSNode/initial_environment.json の初期化
vsnode_dir_path = os.getenv("PROJECT_PATH") + "/VSNode/"
initial_environment_file = vsnode_dir_path + "initial_environment.json"
port_array = {
    "ports": []
}
with open(initial_environment_file, 'w') as file:
    json.dump(port_array, file, indent=4)

# 左下からスタートし，右へ進んでいく
# 端まで到達したら一段上へ
while neLat <= MAX_LAT:
    swLon = MIN_LON
    neLon = swLon + lineStep
    label_lon = 0
    while neLon <= MAX_LON:
        if PSINK_NUM_PER_AREA >= 1:
            i = 0
            while i < PSINK_NUM_PER_AREA:
                for pntype in range(len(pn_types)):
                    vsnode_psnode_relation[pn_types[pntype]] = []
                psink_label = "PS" + str(psink_index)
                j = 0
                while j < len(pn_types):
                    data["psnodes"].append({"psnode":{}, "vsnode":{}})

                    # PSNode情報の追加
                    parea_label = "PA" + str(label_lat) + ":" + str(label_lon)
                    psnode_label = "PSN" + str(id_index)
                    psnode_id = str(int(0b0010 << 60) + id_index)
                    vsnode_id = str(int(0b1000 << 60) + id_index)
                    pnode_type = pn_types[j]
                    vnode_module = os.getenv("PROJECT_PATH") + "/MECServer/VSNode/main"
                    psnode_port = int(os.getenv("PSNODE_BASE_PORT")) + id_index
                    vsnode_port = int(os.getenv("VSNODE_BASE_PORT")) + id_index
                    vnode_socket_address = IP_ADDRESS + ":" + str(vsnode_port)
                    psnode_lat = random.uniform(swLat, neLat)
                    psnode_lon = random.uniform(swLon, neLon)
                    capability = capabilities[pnode_type]
                    credential = "YES"
                    session_key = generate_random_string(10)
                    psnode_description = "Description:" + psnode_label
                    psnode_dict = {
                        "property-label": "PSNode",
                        "relation-label": {
                            "PSink": psink_label,
                            "PArea": parea_label,
                        },
                        "data-property": {
                            "Label": psnode_label,
                            "PNodeID": psnode_id,
                            "PNodeType": pnode_type,
                            "VNodeModule": vnode_module,
                            "SocketAddress": "",
                            "Position": [round(psnode_lat, 4), round(psnode_lon, 4)],
                            "Capability": capability,
                            "Credential": credential,
                            "SessionKey": session_key,
                            "Description": psnode_description
                        },
                        "object-property": [
                            {
                                "from": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "to": {
                                    "property-label": "PSink",
                                    "data-property": "Label",
                                    "value": psink_label
                                },
                                "type": "isConnectedTo"
                            },
                            {
                                "from": {
                                    "property-label": "PSink",
                                    "data-property": "Label",
                                    "value": psink_label
                                },
                                "to": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "type": "aggregates"
                            },
                            {
                                "from": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "to": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "type": "contains"
                            },
                            {
                                "from": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "to": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "type": "isInstalledIn"
                            }
                        ]
                    }
                    data["psnodes"][-1]["psnode"] = psnode_dict

                    # PSNode/initial_environment.json に初期環境に配置されるPSNodeのポート番号を格納
                    initial_environment_file = psnode_dir_path + "initial_environment.json"
                    with open(initial_environment_file, 'r') as file:
                        ports_data = json.load(file)
                    ports_data["ports"].append(psnode_port)
                    with open(initial_environment_file, 'w') as file:
                        json.dump(ports_data, file, indent=4)

                    # VSNode情報の追加
                    parea_label = "PA" + str(label_lat) + ":" + str(label_lon)
                    vsnode_label = "VSN" + str(id_index)
                    vsnode_description = "Description:" + vsnode_label
                    vsnode_dict = {
                        "property-label": "VSNode",
                        "relation-label": {
                            "PArea": parea_label
                        },
                        "data-property": {
                            "Label": vsnode_label,
                            "VNodeID": vsnode_id,
                            "SocketAddress": vnode_socket_address,
                            "SoftwareModule": vnode_module,
                            "Description": vsnode_description
                        },
                        "object-property": [
                            {
                                "from": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "to": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "type": "isVirtualizedBy"
                            },
                            {
                                "from": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "to": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "type": "isPhysicalizedBy"
                            },
                            {
                                "from": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "to": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "type": "isInstalledIn"
                            },
                            {
                                "from": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "to": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "type": "contains"
                            }
                        ]
                    }
                    data["psnodes"][-1]["vsnode"] = vsnode_dict

                    # VSNode/initial_environment.json に初期環境に配置されるVSNodeのポート番号を格納
                    initial_environment_file = vsnode_dir_path + "initial_environment.json"
                    with open(initial_environment_file, 'r') as file:
                        ports_data = json.load(file)
                    ports_data["ports"].append(vsnode_port)
                    with open(initial_environment_file, 'w') as file:
                        json.dump(ports_data, file, indent=4)

                    id_index += 1
                    j += 1
                psink_index += 1
                i += 1
        else:
            interval = int(1 / PSINK_NUM_PER_AREA)
            if label_lon % interval == 0:
                for pntype in range(len(pn_types)):
                    vsnode_psnode_relation[pn_types[pntype]] = []
                psink_label = "PS" + str(psink_index)
                j = 0
                while j < len(pn_types):
                    data["psnodes"].append({"psnode":{}, "vsnode":{}})

                    # PSNode情報の追加
                    parea_label = "PA" + str(label_lat) + ":" + str(label_lon)
                    psnode_label = "PSN" + str(id_index)
                    psnode_id = str(int(0b0010 << 60) + id_index)
                    vsnode_id = str(int(0b1000 << 60) + id_index)
                    pnode_type = pn_types[j]
                    vnode_module = os.getenv("PROJECT_PATH") + "/MECServer/VSNode/main"
                    psnode_port = int(os.getenv("PSNODE_BASE_PORT")) + id_index
                    vsnode_port = int(os.getenv("VSNODE_BASE_PORT")) + id_index
                    vnode_socket_address = IP_ADDRESS + ":" + str(port)
                    psnode_lat = random.uniform(swLat, neLat)
                    psnode_lon = random.uniform(swLon, neLon)
                    capability = capabilities[pnode_type]
                    credential = "YES"
                    session_key = generate_random_string(10)
                    psnode_description = "Description:" + psnode_label
                    psnode_dict = {
                        "property-label": "PSNode",
                        "relation-label": {
                            "PSink": psink_label,
                            "PArea": parea_label,
                        },
                        "data-property": {
                            "Label": psnode_label,
                            "PNodeID": psnode_id,
                            "PNodeType": pnode_type,
                            "VNodeModule": vnode_module,
                            "SocketAddress": "",
                            "Position": [round(psnode_lat, 4), round(psnode_lon, 4)],
                            "Capability": capability,
                            "Credential": credential,
                            "SessionKey": session_key,
                            "Description": psnode_description
                        },
                        "object-property": [
                            {
                                "from": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "to": {
                                    "property-label": "PSink",
                                    "data-property": "Label",
                                    "value": psink_label
                                },
                                "type": "isConnectedTo"
                            },
                            {
                                "from": {
                                    "property-label": "PSink",
                                    "data-property": "Label",
                                    "value": psink_label
                                },
                                "to": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "type": "aggregates"
                            },
                            {
                                "from": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "to": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "type": "contains"
                            },
                            {
                                "from": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "to": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "type": "isInstalledIn"
                            }
                        ]
                    }
                    data["psnodes"][-1]["psnode"] = psnode_dict

                    # PSNode/initial_environment.json に初期環境に配置されるPSNodeのポート番号を格納
                    initial_environment_file = psnode_dir_path + "initial_environment.json"
                    with open(initial_environment_file, 'r') as file:
                        ports_data = json.load(file)
                    ports_data["ports"].append(psnode_port)
                    with open(initial_environment_file, 'w') as file:
                        json.dump(ports_data, file, indent=4)

                    # VSNode情報の追加
                    parea_label = "PA" + str(label_lat) + ":" + str(label_lon)
                    vsnode_label = "VSN" + str(id_index)
                    vsnode_description = "Description:" + vsnode_label
                    vsnode_dict = {
                        "property-label": "VSNode",
                        "relation-label": {
                            "PArea": parea_label
                        },
                        "data-property": {
                            "Label": vsnode_label,
                            "VNodeID": vsnode_id,
                            "SocketAddress": vnode_socket_address,
                            "SoftwareModule": vnode_module,
                            "Description": vsnode_description
                        },
                        "object-property": [
                            {
                                "from": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "to": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "type": "isVirtualizedBy"
                            },
                            {
                                "from": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "to": {
                                    "property-label": "PSNode",
                                    "data-property": "Label",
                                    "value": psnode_label
                                },
                                "type": "isPhysicalizedBy"
                            },
                            {
                                "from": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "to": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "type": "isInstalledIn"
                            },
                            {
                                "from": {
                                    "property-label": "PArea",
                                    "data-property": "Label",
                                    "value": parea_label
                                },
                                "to": {
                                    "property-label": "VSNode",
                                    "data-property": "Label",
                                    "value": vsnode_label
                                },
                                "type": "contains"
                            }
                        ]
                    }
                    data["psnodes"][-1]["vsnode"] = vsnode_dict

                    # VSNode/initial_environment.json に初期環境に配置されるVSNodeのポート番号を格納
                    initial_environment_file = vsnode_dir_path + "initial_environment.json"
                    with open(initial_environment_file, 'r') as file:
                        ports_data = json.load(file)
                    ports_data["ports"].append(vsnode_port)
                    with open(initial_environment_file, 'w') as file:
                        json.dump(ports_data, file, indent=4)

                    id_index += 1
                    j += 1
                psink_index += 1
        label_lon += 1
        swLon = ((swLon*forint) + (lineStep*forint)) / forint
        neLon = ((neLon*forint) + (lineStep*forint)) / forint
    label_lat += 1
    swLat = ((swLat*forint) + (lineStep*forint)) / forint
    neLat = ((neLat*forint) + (lineStep*forint)) / forint


psnode_json = json_file_path + "config_main_psnode.json"
with open(psnode_json, 'w') as f:
    json.dump(data, f, indent=4)