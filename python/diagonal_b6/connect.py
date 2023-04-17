import grpc

from diagonal_b6 import expression

from diagonal_b6 import features_pb2
from diagonal_b6 import api_pb2
from diagonal_b6 import api_pb2_grpc

class Connection:

    def __init__(self, stub, url, token):
        self.stub = stub
        self.url = url
        self.token = token

    def __call__(self, e):
        request = api_pb2.EvaluateRequestProto()
        request.request.CopyFrom(expression.to_node_proto(e))
        return expression.from_node_proto(self.stub.Evaluate(request).result)

def connect_insecure(address):
    channel = grpc.insecure_channel(address)
    return Connection(api_pb2_grpc.B6Stub(channel), _url(address, secure=False), "local")

def _url(address, secure=False):
    if address.find(":") > 0:
        host, port = address.split(":")
        address = host + ":" + str(int(port)-1)
    if secure:
        return "https://" + address
    return "http://" + address