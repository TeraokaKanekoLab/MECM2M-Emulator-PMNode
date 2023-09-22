import json
from dotenv import load_dotenv
import os
import math
import random
from datetime import datetime
import socket
import ipaddress

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
PSINK_IN_PMNODE_BASE_PORT = 5000
VPOINT_NUM_PER_AREA = 1
PSINK_NUM_PER_VPOINT = 1
PMNODE_NUM_PER_PSINK = 1
PSINK_NUM_PER_PMNODE = 1
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

data = {"psinks":[]}

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

#PSinkごとのPSNodeの番号
pmnode_num = 0
psnode_num = 0
port_num = 0

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
                #pmnode_num = 0
                k = 0
                while k < PMNODE_NUM_PER_PSINK:
                    data["psinks"].append({"psink":[], "vpoint":[]})
                    x = data["psinks"][len(data["psinks"])-1]["psink"]
                    y = data["psinks"][len(data["psinks"])-1]["vpoint"]
                    psink_in_pmnode_covered_area = []
                    psink_in_pmnode_num = 0
                    label_pmnode = "PMN" + str(server_num) + ":" + str(pmnode_num)
                    l = 0
                    while l < PSINK_NUM_PER_PMNODE:
                        #PMNode内のPSink情報の追加
                        label_psink_in_pmnode = "PMNPS" + str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_num)
                        psink_in_pmnode_id = "PMNPS" + str(server_num) + "_" + str(pmnode_num) + "_" + str(psink_in_pmnode_num)
                        position_pmnode = search_strings_in_file(pmnode_file_path, label_pmnode, "Position")
                        psink_in_pmnode_lat = round(float(position_pmnode[0][:-1]), 4)
                        psink_in_pmnode_lon = round(float(position_pmnode[1]), 4)
                        psink_in_pmnode_dict = {
                            "property-label": "PSink",
                            "relation-label": {
                                "PMNode": label_pmnode
                            },
                            "data-property": {
                                "Label": label_psink_in_pmnode,
                                "ServingIPv6Address": "",#適当なサブネットマスクを生成する
                                "PSinkID": psink_in_pmnode_id,
                                "Position": [psink_in_pmnode_lat, psink_in_pmnode_lon],#pmnodeのconfigから撮ってくる
                                "Description": "PMNodePSink" + label_psink_in_pmnode
                            },
                            "object-property": [
                            
                            ]
                        }
                        x.append(psink_in_pmnode_dict)
                        psink_in_pmnode_covered_area.append(psink_in_pmnode_num)
                        psink_in_pmnode_num += 1
                        l += 1
                    #VPoint情報の追加
                    if PSINK_NUM_PER_VPOINT == 1:
                        label_vpoint_in_pmnode = "PMNVP" + str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_covered_area[0])
                        vpoint_in_pmnode_id = "PMNVP" + str(server_num) + "_" + str(pmnode_num) + "_" + str(psink_in_pmnode_covered_area[0])
                    else:
                        label_vpoint_in_pmnode = "PMNVP" + str(server_num) + ":" + str(pmnode_num) + ":" + str(psink_in_pmnode_covered_area[0]) + "-" + str(psink_in_pmnode_covered_area[-1])
                        vpoint_in_pmnode_id = "PMNVP" + str(server_num) + "_" + str(pmnode_num) + "_" + str(psink_in_pmnode_covered_area[0]) + "-" + str(psink_in_pmnode_covered_area[-1])
                    port = PSINK_IN_PMNODE_BASE_PORT + port_num
                    vpoint_in_pmnode_dict = {
                        "property-label": "VPoint",
                        "data-property": {
                            "Label": label_vpoint_in_pmnode,
                            "VPointID": vpoint_in_pmnode_id,
                            "Port": str(port),
                            "Description": "VMNodeVPoint" + label_vpoint_in_pmnode
                        },
                        "object-property": [
                        
                        ]
                    }
                    y.append(vpoint_in_pmnode_dict)
                    for i in psink_in_pmnode_covered_area:
                        label_psink_in_pmnode_for_vpoint_in_pmnode = "PMNPS" + str(server_num) + ":" + str(pmnode_num) + ":" + str(i)
                        isComposedOf_object_property = {
                            "from": {
                                "property-label": "VPoint",
                                "data-property": "Label",
                                "value": label_vpoint_in_pmnode
                            },
                            "to": {
                                "property-label": "PSink",
                                "data-property": "Label",
                                "value": label_psink_in_pmnode_for_vpoint_in_pmnode
                            },
                            "type": "isComposedOf"
                        }
                        isVirtualizedWith_object_property = {
                            "from": {
                                "property-label": "PSink",
                                "data-property": "Label",
                                "value": label_psink_in_pmnode_for_vpoint_in_pmnode
                            },
                            "to": {
                                "property-label": "VPoint",
                                "data-property": "Label",
                                "value": label_vpoint_in_pmnode
                            },
                            "type": "isVirtualizedWith"
                        }
                        y[0]["object-property"].append(isComposedOf_object_property)
                        y[0]["object-property"].append(isVirtualizedWith_object_property)
                    port_num += 1
                    pmnode_num += 1
                    k += 1
                j += 1
                if j <= PSINK_NUM_PER_VPOINT:
                    break
            i += 1
            if i <= VPOINT_NUM_PER_AREA:
                break
        label_lon += 1
        swLon = ((swLon*forint) + (lineStep*forint)) / forint
        neLon = ((neLon*forint) + (lineStep*forint)) / forint
    label_lat += 1
    swLat = ((swLat*forint) + (lineStep*forint)) / forint
    neLat = ((neLat*forint) + (lineStep*forint)) / forint
                
psink_in_pmnode_json = json_file_path + "/config_main_psink_in_pmnode.json"
with open(psink_in_pmnode_json, 'w') as f:
    json.dump(data, f, indent=4)