import json
from dotenv import load_dotenv
import os
import random
import argparse

load_dotenv()

config_main_psnode_path = os.getenv("PROJECT_PATH") + "/setup/GraphDB/config/config_main_psnode.json"
with open(config_main_psnode_path, 'r') as file:
    psnode_data = json.load(file)

sql_file_path = os.getenv("PROJECT_PATH") + "/setup/SensingDB/insert_file.sql"
# 初期化
with open(sql_file_path, 'w') as sql_file:
    sql_file.write("")

with open(sql_file_path, 'a') as sql_file:
    insert_command = "INSERT INTO sensordata(PNodeID, Capability, Timestamp, Value, PSinkID, Lat, Lon) VALUES\n"
    sql_file.write(insert_command)

psnode_array = psnode_data["psnodes"]
for minute in range(5):
    for second in range(60):
        if len(str(second)) == 1:
            second_str = f"0{str(second)}"
        else:
            second_str = str(second)
        for i, psnode in enumerate(psnode_array):
            psnode_data_property = psnode["psnode"]["data-property"]
            pnode_id = "\"" + psnode_data_property["PNodeID"] + "\""
            capability = "\"" + psnode_data_property["Capability"] + "\""
            timestamp = "\"" + f"2023-10-31 10:0{str(minute)}:{second_str} +0900 JST" + "\""
            value = round(random.uniform(30, 40), 3)
            psink_id = "\"PSinkID\""
            lat = psnode_data_property["Position"][0]
            lon = psnode_data_property["Position"][1]
            if minute == 4 and second == 59 and i == len(psnode_array)-1:
                insert_line = f"({pnode_id}, {capability}, {timestamp}, {value}, {psink_id}, {lat}, {lon})\n"
            else:
                insert_line = f"({pnode_id}, {capability}, {timestamp}, {value}, {psink_id}, {lat}, {lon}),\n"
            with open(sql_file_path, 'a') as sql_file:
                sql_file.write(insert_line)