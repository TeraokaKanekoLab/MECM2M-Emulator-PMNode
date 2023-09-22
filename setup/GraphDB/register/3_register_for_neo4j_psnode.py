import json
from py2neo import Graph, Node, Relationship
import sys
from dotenv import load_dotenv
import os

load_dotenv()

local_url = "bolt://localhost:" + os.getenv("NEO4J_LOCAL_PORT_PYTHON")
local_graph = Graph(local_url, auth=(os.getenv("NEO4J_USERNAME"), os.getenv("NEO4J_LOCAL_PASSWORD")))
local_tx = local_graph.begin()

args = sys.argv
json_file = os.getenv("HOME") + os.getenv("PROJECT_NAME") + "/setup/GraphDB/config/config_main_psnode.json"
with open(json_file) as f:
    data = json.load(f)

for property in data["psnodes"]:
    # PSNode Data Property
    psnode = property["psnode"]
    label_psnode = psnode["property-label"]
    data_property_psnode = psnode["data-property"]
    node_psnode = Node(label_psnode, **data_property_psnode)
    node_psnode.add_label("PNode")
    local_graph.create(node_psnode)
    
    # VSNode Data Property
    vsnode = property["vsnode"]
    label_vsnode = vsnode["property-label"]
    data_property_vsnode = vsnode["data-property"]
    node_vsnode = Node(label_vsnode, **data_property_vsnode)
    node_vsnode.add_label("VNode")
    local_graph.create(node_vsnode)
    
    # PSNode Object Property
    object_properties_psnode = psnode["object-property"]
    for object_property in object_properties_psnode:
        from_node_label = object_property["from"]["property-label"]
        from_node_property = object_property["from"]["data-property"]
        from_node_value = object_property["from"]["value"]
        to_node_label = object_property["to"]["property-label"]
        to_node_property = object_property["to"]["data-property"]
        to_node_value = object_property["to"]["value"]
        rel_type = object_property["type"]
        from_node = local_graph.nodes.match(from_node_label, **{from_node_property: from_node_value}).first()
        to_node = local_graph.nodes.match(to_node_label, **{to_node_property: to_node_value}).first()
        rel = Relationship(from_node, rel_type, to_node)
        local_graph.create(rel)
    
    # VSNode Object Property
    object_properties_vpoint = vsnode["object-property"]
    for object_property in object_properties_vpoint:
        from_node_label = object_property["from"]["property-label"]
        from_node_property = object_property["from"]["data-property"]
        from_node_value = object_property["from"]["value"]
        to_node_label = object_property["to"]["property-label"]
        to_node_property = object_property["to"]["data-property"]
        to_node_value = object_property["to"]["value"]
        rel_type = object_property["type"]
        from_node = local_graph.nodes.match(from_node_label, **{from_node_property: from_node_value}).first()
        to_node = local_graph.nodes.match(to_node_label, **{to_node_property: to_node_value}).first()
        rel = Relationship(from_node, rel_type, to_node)
        local_graph.create(rel)

try:
    local_graph.commit(local_tx)
except:
    print("Cannot Register Data to Local GraphDB")
else:
    print("Success: PSNode and VSNode Instance in Local GraphDB")
    
#indexの付与
def create_index(graph, object, property):
    query = f"CREATE INDEX index_{object}_{property} IF NOT EXISTS FOR (n:{object}) ON (n.{property});"
    graph.run(query)

create_index(local_graph, "PSNode", "Label")
create_index(local_graph, "VSNode", "Label")