import networkx as nx
import numpy as np
import yaml
import random

rs = np.random.RandomState()


def random_radio_net():
    res = {}
    res["fronthaul_link_capacity"] = ["=", rs.uniform(1000000000, 1000000000000), "resource"]
    res["fronthaul_link_latency"] = ["=", rs.uniform(1, 10)]
    res["pRB_amount"] = ["=", float(rs.choice(range(0, 100, 20), 1)[0])]
    res["mac_scheduler"] = ["=", str(rs.choice(["RR", "PF", "EDF"]))]
    res["price"] = ["=", float(rs.uniform(0, 1000))]
    res["type"] = "radio"
    return res


def random_radio_req():
    res = {}
    res["fronthaul_link_capacity"] = [">", rs.uniform(10000000, 10000000000), "resource"]
    res["fronthaul_link_latency"] = ["<", rs.uniform(1, 20)]
    res["pRB_amount"] = [">", float(rs.choice(range(0, 100, 20), 1)[0])]
    res["mac_scheduler"] = ["=", str(rs.choice(["RR", "PF", "EDF"]))]
    return res


def core_net():
    res = {}
    res["computing_capacity"] = ["=", float(rs.choice(range(20, 60))),"resource"]
    res["memory"] = ["=", float(rs.choice(range(20, 100))),"resource"]
    res["bandwidth"] = ["=", float(rs.choice(range(10, 100))),"resource"]
    res["price"] = ["=", float(rs.uniform(0, 1000))]
    res["type"] = "core"
    return res


def core_req():
    res = {}
    res["computing_capacity"] = [">", float(rs.choice(range(1, 6)))]
    res["memory"] = [">", float(rs.choice(range(1, 10)))]
    res["bandwidth"] = [">", float(rs.choice(range(1, 10)))]
    return res


def transport_net():
    res = {}
    res["bandwidth"] = ["=", float(rs.choice(range(1, 100))), "resource"]
    res["latency"] = ["=", float(rs.choice(range(10, 50)))]
    res["price"] = ["=", float(rs.uniform(0, 1000))]
    res["type"] = "transport"
    return res


def transport_req():
    res = {}
    res["bandwidth"] = [">", float(rs.choice(range(1, 10)))]
    res["latency"] = [">", float(rs.choice(range(1, 10)))]
    return res


def net_generator(type):
    if (type == "transport"):
        return transport_net
    elif (type == "core"):
        return core_net
    elif (type == "radio"):
        return random_radio_net
    else:
        raise Exception("error")


def req_generator(type):
    if (type == "transport"):
        return transport_req
    elif (type == "core"):
        return core_req
    elif (type == "radio"):
        return random_radio_req
    else:
        raise Exception("error")
