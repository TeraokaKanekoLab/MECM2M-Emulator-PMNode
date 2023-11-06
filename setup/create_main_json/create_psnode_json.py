import json
from dotenv import load_dotenv
import os
import random
import string

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

# VSNODE_BASE_PORT, PSNODE_BASE_PORT
# IP_ADDRESS
PSNODE_BASE_PORT = int(os.getenv("PSNODE_BASE_PORT"))
VSNODE_BASE_PORT = int(os.getenv("VSNODE_BASE_PORT"))
IP_ADDRESS = os.getenv("IP_ADDRESS")

data = {"psnodes":[]}

# ID用のindex
id_index = 0

# PNTypeをあらかじめ用意
pn_types = ["AirFlowMeter",
            "VacuumSensor",
            "O2Sensor",
            "AFSensor",
            "SlotPositionSensor",
            "CrankPositionSensor",
            "CamPositionSensor",
            "TemperatureSensorForEngineControl",
            "KnockSensor",
            "AccelPositionSensor",
            "SteeringSensor",
            "HeightControlSensor",
            "WheelSpeedSensor",
            "YawRateSensor",
            "OilTemperatureSensor",
            "TorqueSensorForElectronicPowerSteering",
            "AirBagSensor",
            "UltrasonicSensor",
            "TirePressureSensor",
            "RadarSensor",
            "SensingCamera",
            "TouchSensor",
            "AutoAirConditionerSensor",
            "AutoRightSensor",
            "FuelSensor",
            "RainSensor",
            "AirQualitySensor",
            "GyroSensor",
            "AlcoholInterlockSensor"]
capabilities = {"AirFlowMeter":"AFMAirIntakeAmount",
            "VacuumSensor":"VSAirIntakeAmount",
            "O2Sensor":"O2SOxygenConcentration",
            "AFSensor":"AFSOxygenConcentration",
            "SlotPositionSensor":"SPSAccelPosition",
            "CrankPositionSensor":"CrPSEngineRPM",
            "CamPositionSensor":"CaPSEngineRPM",
            "TemperatureSensorForEngineControl":"TemperatureEC",
            "KnockSensor":"Knocking",
            "AccelPositionSensor":"APSAccelPosition",
            "SteeringSensor":"HandleAngle",
            "HeightControlSensor":"HCSDistance",
            "WheelSpeedSensor":"WheelSpeed",
            "YawRateSensor":"RotationalSpeed",
            "OilTemperatureSensor":"OTemperature",
            "TorqueSensorForElectronicPowerSteering":"TorquePower",
            "AirBagSensor":"AirBag",
            "UltrasonicSensor":"Hz",
            "TirePressureSensor":"kPa",
            "RadarSensor":"RSDistance",
            "SensingCamera":"Frame",
            "TouchSensor":"TouchPosition",
            "AutoAirConditionerSensor":"AASTemperature",
            "AutoRightSensor":"Lux",
            "FuelSensor":"Gallon",
            "RainSensor":"Milli",
            "AirQualitySensor":"Gass",
            "GyroSensor":"Direction",
            "AlcoholInterlockSensor":"Alcohol"}

# PSNode/initial_environment.json の初期化
psnode_dir_path = os.getenv("PROJECT_PATH") + "/PSNode/"
psnode_initial_environment_file = psnode_dir_path + "initial_environment.json"
port_array = {
    "ports": []
}
with open(psnode_initial_environment_file, 'w') as file:
    json.dump(port_array, file, indent=4)

# VSNode/initial_environment.json の初期化
vsnode_dir_path = os.getenv("PROJECT_PATH") + "/VSNode/"
vsnode_initial_environment_file = vsnode_dir_path + "initial_environment.json"
port_array = {
    "ports": []
}
with open(vsnode_initial_environment_file, 'w') as file:
    json.dump(port_array, file, indent=4)

# PNTypeの数だけセンサノードを配置する
for i in range(2):
    for pn_type in pn_types:
        data["psnodes"].append({"psnode":{}, "vsnode":{}})

        # PSNode情報の追加
        psnode_label = "PSN" + str(id_index)
        psnode_id = str(int(0b0010 << 60) + id_index)
        vsnode_id = str(int(0b1000 << 60) + id_index)
        pnode_type = pn_type
        vnode_module = os.getenv("PROJECT_PATH") + "/VSNode/main"
        psnode_port = PSNODE_BASE_PORT + id_index
        vsnode_port = VSNODE_BASE_PORT + id_index
        vnode_socket_address = IP_ADDRESS + ":" + str(vsnode_port)
        capability = capabilities[pnode_type]
        credential = "YES"
        session_key = generate_random_string(10)
        psnode_description = "Description:" + psnode_label
        psnode_dict = {
            "property-label": "PSNode",
            "data-property": {
                "Label": psnode_label,
                "PNodeID": psnode_id,
                "PNodeType": pnode_type,
                "VNodeModule": vnode_module,
                "SocketAddress": "",
                "Position": [0.0, 0.0],
                "Capability": capability,
                "Credential": credential,
                "SessionKey": session_key,
                "Description": psnode_description
            },
            "object-property": [
                
            ]
        }
        data["psnodes"][-1]["psnode"] = psnode_dict

        # PSNode/initial_environment.json に初期環境に配置されるPSNodeのポート番号を格納
        with open(psnode_initial_environment_file, 'r') as file:
            ports_data = json.load(file)
        ports_data["ports"].append(psnode_port)
        with open(psnode_initial_environment_file, 'w') as file:
            json.dump(ports_data, file, indent=4)

        # VSNode情報の追加
        vsnode_label = "VSN" + str(id_index)
        vsnode_description = "Description:" + vsnode_label
        vsnode_dict = {
            "property-label": "VSNode",
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
                }
            ]
        }
        data["psnodes"][-1]["vsnode"] = vsnode_dict

        # VSNode/initial_environment.json に初期環境に配置されるVSNodeのポート番号を格納
        with open(vsnode_initial_environment_file, 'r') as file:
            ports_data = json.load(file)
        ports_data["ports"].append(vsnode_port)
        with open(vsnode_initial_environment_file, 'w') as file:
            json.dump(ports_data, file, indent=4)
        
        id_index += 1


psnode_json = json_file_path + "config_main_psnode.json"
with open(psnode_json, 'w') as f:
    json.dump(data, f, indent=4)