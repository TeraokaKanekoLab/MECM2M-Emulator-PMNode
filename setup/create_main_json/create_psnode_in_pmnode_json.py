import json
from dotenv import load_dotenv
import os
import math
import random
from datetime import datetime
import socket
import ipaddress

# PSNode
# ---------------
## Data Property
## * PNodeID
## * PNode Type (デバイスとクラスの対応づけ)
## * VNode Module ID (VNode moduleの実行系ファイル)
## * Socket Address (VNode のソケットファイル，本来はIP:Port)
## * Lat Lon
## * Capability
## * Credential
## * Session Key
## * Description 
# ---------------
## Object Property
## * aggregates (PSink->PNode)
## * isConnectedTo (PNode->PSink)

# VSNode
# ---------------
## Data Property
## * VNodeID
## * Socket Address (VNode のソケットファイル，本来はIP:Port)
## * Software Module (VNode moduleの実行系ファイル)
## * Description
# ---------------
## Object Property
## * isVirtualizedBy (PNode->VNode)
## * isPhysicalizedBy (VNode->PNode)

def generate_random_ipv6():
    random_int = random.getrandbits(128)
    ipv6 = ipaddress.IPv6Address(random_int)
    return str(ipv6)

def search_strings_in_file(file_path, first_string, second_string):
    first_string_found = False
    line_number = 0
    result_lines = []

    with open(file_path, "r") as file:
        lines = file.readlines()
        for i, line in enumerate(lines):
            line_number += 1
            if not first_string_found and first_string in line:
                first_string_found = True

            if first_string_found and second_string in line:
                if i + 1 < len(lines):
                    result_lines.append(lines[i + 1].strip())
                if i + 2 < len(lines):
                    result_lines.append(lines[i + 2].strip())
                break

    return result_lines

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/Main/config/json_files"

pmnode_file_path = os.getenv("PROJECT_PATH") + "/Main/config/json_files/config_main_pmnode.json"

#PSINK_IN_PMNODE_BASE_PORT
#VPOINT_NUM_PER_AREA
#PSINK_NUM_PER_VPOINT
#PMNODE_NUM_PER_PSINK
#PSINK_NUM_PER_PMNODE
#MIN_LAT, MAX_LAT, MIN_LON, MAX_LON
#AREA_WIDTH
#EDGE_SERVER_NUM
PSNODE_IN_PMNODE_BASE_PORT = 6000
VPOINT_NUM_PER_AREA = 1
PSINK_NUM_PER_VPOINT = 1
PMNODE_NUM_PER_PSINK = 1
PSINK_NUM_PER_PMNODE = 1
VSNODE_NUM_PER_PNTYPE = 1
MIN_LAT = 35.530
MAX_LAT = 35.540
MIN_LON = 139.530
MAX_LON = 139.540
AREA_WIDTH = 0.001
EDGE_SERVER_NUM = 3

lineStep = AREA_WIDTH
forint = 1000

area_num = math.ceil(((MAX_LAT-MIN_LAT)/AREA_WIDTH)*((MAX_LON-MIN_LON)/AREA_WIDTH))
area_num_per_server = int(area_num / EDGE_SERVER_NUM)

data = {"psnodes":{"psnode":[], "vsnode":[]}}

#始点となるArea
swLat = MIN_LAT
neLat = swLat + lineStep

#label情報
label_lat = 0
label_lon = 0

#server_counter
server_counter = 0
server_num = 1

#ServerごとのPSinkの番号
psink_in_pmnode_num = 0
psnode_in_pmnode_num = 0

#PSinkごとのPSNodeの番号
pmnode_num = 0
psnode_num = 0
port_num = 0

#PSNodeのIPv6Pref用のインデックス
psnode_ipv6_pref = 1 << 63

#PNTypeをあらかじめ用意
pn_types = ["Temp_Sensor", "Humid_Sensor", "Anemometer"]
capabilities = {"Temp_Sensor":"MaxTemp", "Humid_Sensor":"MaxHumid", "Anemometer":"MaxWind"}
vsnode_psnode_relation = {}
for i in range(len(pn_types)):
    vsnode_psnode_relation[pn_types[i]] = []

#左下からスタートし，右へ進んでいく
#端まで到達したら一段上へ
while neLat <= MAX_LAT:
    swLon = MIN_LON
    neLon = swLon + lineStep
    label_lon = 0
    while neLon <= MAX_LON:
        server_counter += 1
        if server_num > EDGE_SERVER_NUM:
            server_num -= 1
        if server_counter == area_num_per_server*server_num+1 and server_num < EDGE_SERVER_NUM:
            pmnode_num = 0
            server_num += 1
        if (area_num_per_server*(server_num-1)) <= server_counter < (area_num_per_server*server_num):
            label_server = "S" + str(server_num)
        i = 0
        while i < VPOINT_NUM_PER_AREA:
            j = 0
            while j < PSINK_NUM_PER_VPOINT:
                k = 0
                while k < PMNODE_NUM_PER_PSINK:
                    psink_in_pmnode_covered_area = []
                    psink_in_pmnode_num = 0
                    label_pmnode = "PMN" + str(server_num) + ":" + str(pmnode_num)
                    for pntype in range(len(pn_types)):
                        vsnode_psnode_relation[pn_types[pntype]] = []
                    l = 0
                    while l < PSINK_NUM_PER_PMNODE:
                        m = 0
                        label_psink_in_pmnode = "PMNPS" + str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_num)
                        psnode_in_pmnode_num = 0
                        while m < len(pn_types):
                            # 20230426 ここから
                            # PMNode内のPSNodeの情報を追加
                            label_psnode_in_pmnode = "PMNPSN" + str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_num) + ":" + str(psnode_in_pmnode_num)
                            position_pmnode = search_strings_in_file(pmnode_file_path, label_pmnode, "Position")
                            psnode_in_pmnode_lat = round(float(position_pmnode[0][:-1]), 4)
                            psnode_in_pmnode_lon = round(float(position_pmnode[1]), 4)
                            psnode_in_pmnode_socket_file = "/tmp/mecm2m/psnode_" + str(server_num) + "_" + str(pmnode_num) + "_" + str(psink_in_pmnode_num) + "_" + str(psnode_in_pmnode_num) + ".sock"
                            psnode_pn_type = pn_types[m]
                            psnode_in_pmnode_dict = {
                                "property-label": "PSNode",
                                "relation-label": {
                                    "PSink": label_psink_in_pmnode,
                                    "PNodeID": str(psnode_ipv6_pref),
                                    "socket-file": psnode_in_pmnode_socket_file
                                },
                                "data-property": {
                                    "Label": label_psnode_in_pmnode,
                                    "PNodeID": str(psnode_ipv6_pref),
                                    "IPv6Pref": str(psnode_ipv6_pref),
                                    "PNType": psnode_pn_type,
                                    "Position": [round(psnode_in_pmnode_lat, 4), round(psnode_in_pmnode_lon, 4)],
                                    "Capability": capabilities[psnode_pn_type],
                                    "Credential": "YES",
                                    "Description": "PMNodePSNode" + label_psnode_in_pmnode
                                },
                                "object-property": [
                                    {
                                        "from": {
                                            "property-label": "PSNode",
                                            "data-property": "Label",
                                            "value": label_psnode_in_pmnode
                                        },
                                        "to": {
                                            "property-label": "PSink",
                                            "data-property": "Label",
                                            "value": label_psink_in_pmnode
                                        },
                                        "type": "respondsViaDevApi"
                                    },
                                    {
                                        "from": {
                                            "property-label": "PSink",
                                            "data-property": "Label",
                                            "value": label_psink_in_pmnode
                                        },
                                        "to": {
                                            "property-label": "PSNode",
                                            "data-property": "Label",
                                            "value": label_psnode_in_pmnode
                                        },
                                        "type": "requestsViaDevApi"
                                    }
                                ]
                            }
                            data["psnodes"]["psnode"].append(psnode_in_pmnode_dict)
                            vsnode_psnode_relation[psnode_pn_type].append(label_psnode_in_pmnode)
                            psnode_in_pmnode_num += 1
                            psnode_ipv6_pref += 1
                            m += 1
                        psink_in_pmnode_covered_area.append(psink_in_pmnode_num)
                        psink_in_pmnode_num += 1
                        l += 1
                    #VPoint情報の追加
                    if PSINK_NUM_PER_VPOINT == 1:
                        label_vpoint_in_pmnode = str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_covered_area[0])
                        vpoint_in_pmnode_id = str(server_num) + "_" + str(pmnode_num) + "_" + str(psink_in_pmnode_covered_area[0])
                    else:
                        label_vpoint_in_pmnode = str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_covered_area[0]) + "-" + str(psink_in_pmnode_covered_area[-1])
                        vpoint_in_pmnode_id = str(server_num) + "_" + str(pmnode_num) + "_" + str(psink_in_pmnode_covered_area[0]) + "-" + str(psink_in_pmnode_covered_area[-1])
                    #VSNodeの情報の追加
                    a = 0
                    vsnode_num = 0
                    while a < VSNODE_NUM_PER_PNTYPE:
                        b = 0
                        while b < len(pn_types):
                            label_vsnode = "PMNVSN" + label_vpoint_in_pmnode + ":" + str(vsnode_num)
                            vsnode_id = "PMNVSN" + vpoint_in_pmnode_id + "_" + str(vsnode_num)
                            port = PSNODE_IN_PMNODE_BASE_PORT + port_num
                            vsnode_in_pmnode_dict = {
                                "property-label": "VSNode",
                                "data-property": {
                                "Label": label_vsnode,
                                "VNodeID": vsnode_id,
                                "VNType": pn_types[b],
                                "Port": str(port),
                                "Description": "PMNodeVSNode" + label_vsnode
                                },
                                "object-property": [
                                
                                ]
                            }
                            data["psnodes"]["vsnode"].append(vsnode_in_pmnode_dict)
                            for pntype in vsnode_psnode_relation[pn_types[b]]:
                                isComposedOf_object_property = {
                                    "from": {
                                        "property-label": "VSNode",
                                        "data-property": "Label",
                                        "value": label_vsnode
                                    },
                                    "to": {
                                        "property-label": "PSNode",
                                        "data-property": "Label",
                                        "value": pntype
                                    },
                                    "type": "isComposedOf"
                                }
                                isVirtualizedWith_object_property = {
                                    "from": {
                                        "property-label": "PSNode",
                                        "data-property": "Label",
                                        "value": pntype
                                    },
                                    "to": {
                                        "property-label": "VSNode",
                                        "data-property": "Label",
                                        "value": label_vsnode
                                    },
                                    "type": "isVirtualizedWith"
                                }
                                data["psnodes"]["vsnode"][-1]["object-property"].append(isComposedOf_object_property)
                                data["psnodes"]["vsnode"][-1]["object-property"].append(isVirtualizedWith_object_property)
                            b += 1
                        a += 1
                        vsnode_num += 1
                        port_num += 1
                    pmnode_num += 1
                    k += 1
                j += 1
            i += 1
        label_lon += 1
        swLon = ((swLon*forint) + (lineStep*forint)) / forint
        neLon = ((neLon*forint) + (lineStep*forint)) / forint
    label_lat += 1
    swLat = ((swLat*forint) + (lineStep*forint)) / forint
    neLat = ((neLat*forint) + (lineStep*forint)) / forint
                
psnode_in_pmnode_json = json_file_path + "/config_main_psnode_in_pmnode.json"
with open(psnode_in_pmnode_json, 'w') as f:
    json.dump(data, f, indent=4)