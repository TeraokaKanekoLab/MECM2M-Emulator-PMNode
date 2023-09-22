import json
from dotenv import load_dotenv
import os
import random
import math

# PArea
# ---------------
## Data Property
## * PAreaID
## * SW Lat Lon
## * NE Lat Lon
## * Description
# ---------------
## Object Property
## * isEastOf (PArea)
## * isWestOf (PArea)
## * isSouthOf (PArea)
## * isNorthOf (PArea)

load_dotenv()
json_file_path = os.getenv("PROJECT_PATH") + "/setup/GraphDB/config/"

# MIN_LAT, MAX_LAT, MIN_LON, MAX_LON
# AREA_WIDTH
MIN_LAT = float(os.getenv("MIN_LAT"))
MAX_LAT = float(os.getenv("MAX_LAT"))
MIN_LON = float(os.getenv("MIN_LON"))
MAX_LON = float(os.getenv("MAX_LON"))
AREA_WIDTH = float(os.getenv("AREA_WIDTH"))

lineStep = AREA_WIDTH
forint = 1000

data = {"areas":{"parea":[]}}

# 始点となるArea
swLat = MIN_LAT
neLat = swLat + lineStep

# label情報
label_lat = 0
label_lon = 0

# PAreaIDのインデックス
parea_id_index = 0

# 左下からスタートし，右へ進んでいく
# 端まで到達したら一段上へ
while neLat <= MAX_LAT:
    swLon = MIN_LON
    neLon = swLon + lineStep
    label_lon = 0
    while neLon <= MAX_LON:
        parea = data["areas"]["parea"]
        parea_label = "PA" + str(label_lat) + ":" + str(label_lon)
        # parea_id = str(int(swLat*1000)) + "0" + str(int(swLon*1000)) + "0"
        parea_id = str(int(0b0000 << 60) + parea_id_index)
        parea_description = "Description:" + parea_label

        parea_dict = {
            "property-label": "PArea",
            "data-property": {
                "Label": parea_label,
                "PAreaID": parea_id,
                "SW": [round(swLat, 3), round(swLon, 3)],
                "NE": [round(neLat, 3), round(neLon, 3)],
                "Description": parea_description
            },
            "object-property": [
            
            ]
        }
        object_properties = parea_dict["object-property"]
        if label_lat > 0:
            # isNorthOf, isSouthOf
            isSouth_label = "PA" + str(label_lat-1) + ":" + str(label_lon)
            isNorthOf_object_property = {
                "from": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": parea_label
                },
                "to": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": isSouth_label
                },
                "type": "isNorthOf"
            }
            isSouthOf_object_property = {
                "from": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": isSouth_label
                },
                "to": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": parea_label
                },
                "type": "isSouthOf"
            }
            object_properties.append(isNorthOf_object_property)
            object_properties.append(isSouthOf_object_property)
        if label_lon > 0:
            # isEastOf, isWestOf
            isWest_label = "PA" + str(label_lat) + ":" + str(label_lon-1)
            isEastOf_object_property = {
                "from": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": parea_label
                },
                "to": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": isWest_label
                },
                "type": "isEastOf"
            }
            isWestOf_object_property = {
                "from": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": isWest_label
                },
                "to": {
                    "property-label": "PArea",
                    "data-property": "Label",
                    "value": parea_label
                },
                "type": "isWestOf"
            }
            object_properties.append(isEastOf_object_property)
            object_properties.append(isWestOf_object_property)

        parea.append(parea_dict)
        parea_id_index += 1
        label_lon += 1
        swLon = ((swLon*forint) + (lineStep*forint)) / forint
        neLon = ((neLon*forint) + (lineStep*forint)) / forint
    label_lat += 1
    swLat = ((swLat*forint) + (lineStep*forint)) / forint
    neLat = ((neLat*forint) + (lineStep*forint)) / forint
    
area_json = json_file_path + "config_main_area.json"
with open(area_json, 'w') as f:
    json.dump(data, f, indent=4)