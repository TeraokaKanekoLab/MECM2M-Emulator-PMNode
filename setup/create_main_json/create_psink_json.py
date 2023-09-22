import json
from dotenv import load_dotenv
import os
import math
import random
import ipaddress
import glob

# PSink
# ---------------
## Data Property
## * PSinkID
## * PSink Type (デバイスとクラスの対応づけ)
## * Socket Address (PSinkへのアクセス用)
## * Serving IPv6 Prefix (デバイスの接続用subnet)
## * MEC IPv6 Address (MEC ServerのIPv6アドレス)
## * Lat Lon
## * Description 
# ---------------
## Object Property
## * contains (PArea->PSink)
## * isInstalledIn (PSink->PArea)

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/setup/GraphDB/config/"

# PSINK_NUM_PER_AREA
# MIN_LAT, MAX_LAT, MIN_LON, MAX_LON
# AREA_WIDTH
# IP_ADDRESS
PSINK_NUM_PER_AREA = float(os.getenv("PSINK_NUM_PER_AREA"))
MIN_LAT = float(os.getenv("MIN_LAT"))
MAX_LAT = float(os.getenv("MAX_LAT"))
MIN_LON = float(os.getenv("MIN_LON"))
MAX_LON = float(os.getenv("MAX_LON"))
AREA_WIDTH = float(os.getenv("AREA_WIDTH"))
IP_ADDRESS = os.getenv("IP_ADDRESS")

lineStep = AREA_WIDTH
forint = 1000

data = {"psinks":[]}

# 始点となるArea
swLat = MIN_LAT
neLat = swLat + lineStep

# label情報
label_lat = 0
label_lon = 0

# ID用のindex
psink_id_index = 0

# 左下からスタートし，右へ進んでいく
# 端まで到達したら一段上へ
while neLat <= MAX_LAT:
    swLon = MIN_LON
    neLon = swLon + lineStep
    label_lon = 0
    while neLon <= MAX_LON:
        parea_label = "PA" + str(label_lat) + ":" + str(label_lon)
        # PSINK_NUM_PER_AREA が 1以上か1より小さいかで分岐
        if PSINK_NUM_PER_AREA >= 1:
            i = 0
            while i < PSINK_NUM_PER_AREA:
                psinks = data["psinks"]

                # PSink情報の追加
                psink_label = "PS" + str(psink_id_index)
                psink_id = str(int(0b0001 << 60) + psink_id_index)
                psink_type = "Router"
                socket_address = ""
                random_ipv6 = ipaddress.IPv6Address(random.randint(0, 2**128 - 1))
                serving_ipv6_prefix = str(ipaddress.IPv6Network((random_ipv6, 64), strict=False))
                mec_ipv6_address = IP_ADDRESS
                psink_lat = random.uniform(swLat, neLat)
                psink_lon = random.uniform(swLon, neLon)
                psink_description = "Description:" + psink_label
                psink_dict = {
                    "property-label": "PSink",
                    "relation-label": {
                        "PArea": parea_label
                    },
                    "data-property": {
                        "Label": psink_label,
                        "PSinkID": psink_id,
                        "PSinkType": psink_type,
                        "SocketAddress": socket_address,
                        "ServingIPv6Prefix": serving_ipv6_prefix,   # 適当なサブネットマスクを生成する
                        "MECIPv6Address": mec_ipv6_address,
                        "Position": [round(psink_lat, 4), round(psink_lon, 4)],
                        "Description": psink_description
                    },
                    "object-property": [
                        {
                            "from": {
                                "property-label": "PSink",
                                "data-property": "Label",
                                "value": psink_label
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
                                "property-label": "PSink",
                                "data-property": "Label",
                                "value": psink_label
                            },
                            "type": "contains"
                        }
                    ]
                }
                psinks.append(psink_dict)
                
                psink_id_index += 1
                i += 1
        # PSINK_NUM_PER_AREA が1より小さい場合，〜〜Areaに1個PSinkを設置
        else:
            interval = int(1 / PSINK_NUM_PER_AREA)
            if label_lon % interval == 0:
                psinks = data["psinks"]

                # PSink情報の追加
                psink_label = "PS" + str(psink_id_index)
                psink_id = str(int(0b0001 << 60) + psink_id_index)
                psink_type = "Router"
                socket_address = ""
                random_ipv6 = ipaddress.IPv6Address(random.randint(0, 2**128 - 1))
                serving_ipv6_prefix = str(ipaddress.IPv6Network((random_ipv6, 64), strict=False))
                mec_ipv6_address = IP_ADDRESS
                psink_lat = random.uniform(swLat, neLat)
                psink_lon = random.uniform(swLon, neLon)
                psink_description = "Description:" + psink_label
                psink_dict = {
                    "property-label": "PSink",
                    "relation-label": {
                        "PArea": parea_label
                    },
                    "data-property": {
                        "Label": psink_label,
                        "PSinkID": psink_id,
                        "PSinkType": psink_type,
                        "SocketAddress": socket_address,
                        "ServingIPv6Prefix": serving_ipv6_prefix,   # 適当なサブネットマスクを生成する
                        "MECIPv6Address": mec_ipv6_address,
                        "Position": [round(psink_lat, 4), round(psink_lon, 4)],
                        "Description": psink_description
                    },
                    "object-property": [
                        {
                            "from": {
                                "property-label": "PSink",
                                "data-property": "Label",
                                "value": psink_label
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
                                "property-label": "PSink",
                                "data-property": "Label",
                                "value": psink_label
                            },
                            "type": "contains"
                        }
                    ]
                }
                psinks.append(psink_dict)
                
                psink_id_index += 1
        label_lon += 1
        swLon = ((swLon*forint) + (lineStep*forint)) / forint
        neLon = ((neLon*forint) + (lineStep*forint)) / forint
    label_lat += 1
    swLat = ((swLat*forint) + (lineStep*forint)) / forint
    neLat = ((neLat*forint) + (lineStep*forint)) / forint

psink_json = json_file_path + "config_main_psink.json"
with open(psink_json, 'w') as f:
    json.dump(data, f, indent=4)