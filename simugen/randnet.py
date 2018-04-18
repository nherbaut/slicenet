import networkx as nx
import numpy as np
import yaml
import random
from randstuff import *

generators_net = [random_radio_net, core_net, transport_net]

generators_req = [random_radio_net, core_net, transport_net]

random_graph = nx.generators.random_graphs.powerlaw_cluster_graph(20, 2, 0.3)
for s, e in random_graph.edges:
    generator = rs.choice(generators_net)
    for k, v in generator().items():
        random_graph[s][e][k] = v

net = {}
net["edges"] = [
    {"id": "%s/%s" % (s, e), "node1": s, "node2": e, "attrs": {k: v for k, v in attr.items() if k != "type"},
     "type": attr.get("type", None)} for s, e, attr in
    list(random_graph.edges(data=True))]

with open("net.yaml", "w") as f:
    f.write(yaml.dump(net))


def cleanup(x, g):
    a_type = g[x[0]][x[1]].pop("type")

    g.remove_edge(x[0],
                  x[1])

    g.add_edge(x[0], x[1], **req_generator(a_type)())


edges = list(map(lambda x: cleanup(x, random_graph), list(random_graph.edges)))

done = False
while not done:
    n1, n2 = rs.choice(random_graph.nodes, 2)

    for i, apath in enumerate(list(nx.all_simple_paths(random_graph, n1, n2))[0:10]):
        apath = list(zip(apath, apath[1:]))
        with open("request%d.yaml" % i, "w") as f:
            net = {}
            net["edges"] = [
                {"node1": s, "node2": e, "attrs": attr} for s, e, attr in
                list(random_graph.edges(data=True)) if (s, e) in apath or (e, s) in apath]

            f.write(yaml.dump(net))
            done = True
