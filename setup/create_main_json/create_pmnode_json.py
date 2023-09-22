import json
from dotenv import load_dotenv
import os
import math
import random
from datetime import datetime
import socket
import ipaddress

# PMNode
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
## * Home IPv6 Address
## * Update Time
## * Velocity
## * Acceleration
## * Direction
# ---------------
## Object Property
## * aggregates (PSink->PNode)
## * isConnectedTo (PNode->PSink)

# VMNode
# ---------------
## Data Property
## * VNodeID
## * Socket Address (VNode のソケットファイル，本来はIP:Port)
## * Software Module (VNode moduleの実行系ファイル)
## * Description
## * Foreign Socket Address (VMNodeF のソケットファイル，本来はIP:Port)
## * Representative Socket Address (VMNodeR のソケットファイル，本来はIP:Port)
# ---------------
## Object Property
## * isVirtualizedBy (PNode->VNode)
## * isPhysicalizedBy (VNode->PNode)

def generate_random_ipv6():
    random_int = random.getrandbits(128)
    ipv6 = ipaddress.IPv6Address(random_int)
    return str(ipv6)

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/Main/config/json_files"

#VMNODEH_BASE_PORT
#VPOINT_NUM_PER_AREA
#PSINK_NUM_PER_VPOINT
#PMNODE_NUM_PER_PSINK
#MIN_LAT, MAX_LAT, MIN_LON, MAX_LON
#AREA_WIDTH
#EDGE_SERVER_NUM
VMNODEH_BASE_PORT = 4000
VPOINT_NUM_PER_AREA = 1
PSINK_NUM_PER_VPOINT = 1
PMNODE_NUM_PER_PSINK = 1
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

data = {"pmnodes":[]}

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
psink_num = 0
pmnode_num = 0

#PSinkごとのPSNodeの番号
psnode_num = 0
port_num = 0

#PMNodeのIPv6Pref用のインデックス
pmnode_ipv6_pref = 0

#PNTypeをあらかじめ用意
pn_types = ["Toyota", "Matsuda", "Nissan"]
capabilities = {"Toyota":"Prius", "Matsuda":"Road-Star", "Nissan":"Selena"}

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
            psink_num = 0
            server_num += 1
        if (area_num_per_server*(server_num-1)) <= server_counter < (area_num_per_server*server_num):
            label_server = "S" + str(server_num)
        i = 0
        while i < VPOINT_NUM_PER_AREA:
            j = 0
            while j < PSINK_NUM_PER_VPOINT:
                data["pmnodes"].append({"pmnode":[], "mserver":[], "vmnodeh":[]})
                x = data["pmnodes"][len(data["pmnodes"])-1]["pmnode"]
                y = data["pmnodes"][len(data["pmnodes"])-1]["mserver"]
                z = data["pmnodes"][len(data["pmnodes"])-1]["vmnodeh"]
                pmnode_covered_area = []
                label_home_server = label_server
                k = 0
                while k < PMNODE_NUM_PER_PSINK:
                    #PMNodeの情報の追加
                    label_pmnode = "PMN" + str(server_num) + ":" + str(pmnode_num)
                    #pmnode_id = "PMN" + str(label_home_server) + ":" + str(pmnode_num)
                    pmnode_lat = random.uniform(swLat, neLat)
                    pmnode_lon = random.uniform(swLon, neLon)
                    pmnode_socket_file = "/tmp/mecm2m/pmnode_" + str(server_num) + "_" + str(pmnode_num)
                    pmnode_pn_type = random.choice(pn_types)
                    pmnode_dict = {
                        "property-label": "PMNode",
                        "relation-label": {
                            "HomeServer": label_home_server,
                            "PNodeID": str(pmnode_ipv6_pref)
                        },
                        "data-property": {
                            "Label": label_pmnode,
                            "PNodeID": str(pmnode_ipv6_pref),
                            "IPv6Pref": str(pmnode_ipv6_pref),
                            "PNType": pmnode_pn_type,
                            "Position": [round(pmnode_lat, 4), round(pmnode_lon, 4)],
                            "Capability": capabilities[pmnode_pn_type],
                            "Credential": "YES",
                            "Description": "PMNode" + label_pmnode,
                            "HomeIPv6Pref": "",
                            "UpdateTime": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                            "Velocity": 0,
                            "Accelaration": 0,
                            "Direction": ""
                        },
                        "object-property": [
                        
                        ]
                    }
                    label_mserver = "MS" + str(server_num) + ":" + str(pmnode_num)
                    mserver_dict = {
                        "property-label": "MServer",
                        "data-property": {
                            "Label": label_mserver,
                            "IPv6Address": generate_random_ipv6(),
                            "ServedIPv6Pref": str(pmnode_num),
                            "Description": "MServer" + label_mserver
                        },
                        "object-property": [
                            {
                                "from": {
                                    "property-label": "MServer",
                                    "data-property": "Label",
                                    "value": label_mserver
                                },
                                "to": {
                                    "property-label": "Server",
                                    "data-property": "Label",
                                    "value": label_home_server
                                },
                                "type": "isLowerOf"
                            },
                            {
                                "from": {
                                    "property-label": "Server",
                                    "data-property": "Label",
                                    "value": label_home_server
                                },
                                "to": {
                                    "property-label": "MServer",
                                    "data-property": "Label",
                                    "value": label_mserver
                                },
                                "type": "isUpperOf"
                            },
                            {
                                "from": {
                                    "property-label": "MServer",
                                    "data-property": "Label",
                                    "value": label_mserver
                                },
                                "to": {
                                    "property-label": "PMNode",
                                    "data-property": "Label",
                                    "value": label_pmnode
                                },
                                "type": "isRegardedAs"
                            },
                            {
                                "from": {
                                    "property-label": "PMNode",
                                    "data-property": "Label",
                                    "value": label_pmnode
                                },
                                "to": {
                                    "property-label": "MServer",
                                    "data-property": "Label",
                                    "value": label_mserver
                                },
                                "type": "isRegardedAs"
                            }
                        ]
                    }
                    x.append(pmnode_dict)
                    y.append(mserver_dict)
                    pmnode_covered_area.append(pmnode_num)
                    pmnode_num += 1
                    pmnode_ipv6_pref += 1
                    k += 1
                #VMNodeH情報の追加
                if PMNODE_NUM_PER_PSINK == 1:
                    label_vmnodeh = "VMNH" + str(server_num) + ":" + str(pmnode_covered_area[0])
                    vmnodeh_id = "VMNH" + str(server_num) + "_" + str(pmnode_covered_area[0])
                else:
                    label_vmnodeh = "VMNH" + str(server_num) + ":" + str(pmnode_covered_area[0]) + "-" + str(pmnode_covered_area[-1])
                    vmnodeh_id = "VMNH" + str(server_num) + "_" + str(pmnode_covered_area[0]) + "-" + str(pmnode_covered_area[-1])
                port = VMNODEH_BASE_PORT + port_num
                vmnodeh_dict = {
                    "property-label": "VMNodeH",
                    "data-property": {
                        "Label": label_vmnodeh,
                        "VNodeID": vmnodeh_id,
                        "Port": str(port),
                        "Description": "VMNodeH" + label_vmnodeh
                    },
                    "object-property": [
                    
                    ]
                }
                z.append(vmnodeh_dict)
                for i in pmnode_covered_area:
                    label_pmnode_for_vmnodeh = "PMN" + str(server_num) + ":" + str(i)
                    isComposedOf_object_property = {
                        "from": {
                            "property-label": "VMNodeH",
                            "data-property": "Label",
                            "value": label_vmnodeh
                        },
                        "to": {
                            "property-label": "PMNode",
                            "data-property": "Label",
                            "value": label_pmnode_for_vmnodeh
                        },
                        "type": "isComposedOf"
                    }
                    isVirtualizedWith_object_property = {
                        "from": {
                            "property-label": "PMNode",
                            "data-property": "Label",
                            "value": label_pmnode_for_vmnodeh
                        },
                        "to": {
                            "property-label": "VMNodeH",
                            "data-property": "Label",
                            "value": label_vmnodeh
                        },
                        "type": "isVirtualizedWith"
                    }
                    z[0]["object-property"].append(isComposedOf_object_property)
                    z[0]["object-property"].append(isVirtualizedWith_object_property)
                psink_num += 1
                port_num += 1
                j +=1
            i += 1
        label_lon += 1
        swLon = ((swLon*forint) + (lineStep*forint)) / forint
        neLon = ((neLon*forint) + (lineStep*forint)) / forint
    label_lat += 1
    swLat = ((swLat*forint) + (lineStep*forint)) / forint
    neLat = ((neLat*forint) + (lineStep*forint)) / forint

pmnode_json = json_file_path + "/config_main_pmnode.json"
with open(pmnode_json, 'w') as f:
    json.dump(data, f, indent=4)