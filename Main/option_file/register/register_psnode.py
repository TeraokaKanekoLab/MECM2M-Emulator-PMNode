import json
from py2neo import Graph, Node, Relationship
import sys
from dotenv import load_dotenv
import os
import random

load_dotenv()
url = "bolt://localhost:" + os.getenv("NEO4J_PORT_PYTHON")
graph = Graph(url, auth=(os.getenv("NEO4J_USERNAME"), os.getenv("NEO4J_PASSWORD")))

tx = graph.begin()

args = sys.argv
if len(args) > 1:
    json_file = args[1]
else:
    json_file = os.getenv("PROJECT_PATH") + "/Main/option_file/register/register_psnode.json"
with open(json_file) as f:
    data = json.load(f)

psnode = data["psnodes"]["psnode"]

for node in psnode:
    # configファイルから属性要素を取り出す
    data_property = node["data-property"]
    pn_type = data_property["PNType"]
    position = data_property["Position"]
    capability = data_property["Capability"]
    credential = data_property["Credential"]
    # データプロパティの中身を完全体にしたい (Label, PNodeID, IPv6Pref, Description)
    # まずは，設置する場所から，接続するPSinkを探し出す
    query = f"MATCH (n:PSink) WHERE NOT n.Label CONTAINS 'PM' AND n.Position[0] >= {position[0]} AND n.Position[0] < {position[0] + 0.001} AND n.Position[1] >= {position[1]} AND n.Position[1] < {position[1] + 0.001} RETURN n;"
    result_psink = graph.run(query)
    psink_list = list(result_psink)

    random_psink = random.choice(psink_list)
    random_psink_record = random_psink.values()
    random_psink_properties = random_psink_record[0]
    random_psink_label = random_psink_properties["Label"]

    # 接続先予定のPSinkにすでに接続しているPSNodeの連番の最大値を調べる
    random_psink_label_for_psnode = random_psink_label[2:] + ":"
    query = f"MATCH (n:PSNode) WHERE n.Label CONTAINS '{random_psink_label_for_psnode}' AND NOT n.Label CONTAINS 'PM' RETURN n ORDER BY n.Label DESC LIMIT 1;"
    result_psnode = graph.run(query)
    psnode_list = list(result_psnode)

    psnode_value = psnode_list[0].values()
    psnode_properties = psnode_value[0]
    psnode_max_label = psnode_properties["Label"]
    last_occurrence = psnode_max_label.rfind(":")
    new_max_label = int(psnode_max_label[last_occurrence+1:]) + 1
    psnode_label = "PSN" + random_psink_label_for_psnode + str(new_max_label)
    node["data-property"]["Label"] = psnode_label
    
    psnode_description = "PSNode" + psnode_label
    node["data-property"]["Description"] = psnode_description

    # PNodeIDの最大値を調べる
    query = "MATCH (n:PSNode) WHERE NOT n.Label CONTAINS 'PM' WITH max(toInteger(n.PNodeID)) AS maxID MATCH (n:PSNode) WHERE NOT n.Label CONTAINS 'PM' AND toInteger(n.PNodeID) = maxID RETURN n;"
    result_psnode_maxID = graph.run(query)
    psnode_maxID_list = list(result_psnode_maxID)

    psnode_maxID_value = psnode_maxID_list[0].values()
    psnode_maxID = int(psnode_maxID_value[0]["PNodeID"])
    node["data-property"]["PNodeID"] = str(psnode_maxID + 1)
    node["data-property"]["IPv6Pref"] = str(psnode_maxID + 1)
    
    # データプロパティの登録
    regist = Node("PSNode", **data_property)
    regist.add_label("PNode")
    graph.create(regist)

    # オブジェクトプロパティの登録
    psnode_for_rel = graph.nodes.match("PSNode", **{"Label": psnode_label}).first()
    psink_for_rel = graph.nodes.match("PSink", **{"Label": random_psink_label}).first()
    print(psnode_for_rel)
    rel_psink_psnode = Relationship(psnode_for_rel, "respondsViaDevApi", psink_for_rel)
    rel_psnode_psink = Relationship(psink_for_rel, "requestsViaDevApi", psnode_for_rel)
    graph.create(rel_psink_psnode)
    graph.create(rel_psnode_psink)

try:
    graph.commit(tx)
except:
    print("Cannot Register PSNode to Neo4j")
else:
    print("Success: PSNode Instance")
