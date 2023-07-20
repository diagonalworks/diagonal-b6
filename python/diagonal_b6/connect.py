import grpc

from diagonal_b6 import expression
from diagonal_b6 import VERSION

from diagonal_b6 import api_pb2
from diagonal_b6 import api_pb2_grpc

class Connection:

    def __init__(self, stub):
        self.stub = stub

    def __call__(self, e):
        request = api_pb2.EvaluateRequestProto()
        request.version = VERSION
        node = expression.to_node(e)
        request.request.CopyFrom(node.to_node_proto())
        return expression.from_node_proto(self.stub.Evaluate(request).result)

def connect_insecure(address):
    channel = grpc.insecure_channel(address)
    return Connection(api_pb2_grpc.B6Stub(channel))

def connect(address, token, root_certificates=None):
    channel_credentials = grpc.ssl_channel_credentials(root_certificates=root_certificates)
    credentials = grpc.composite_channel_credentials(channel_credentials, grpc.access_token_call_credentials(token))
    channel = grpc.secure_channel(address, credentials)
    return Connection(api_pb2_grpc.B6Stub(channel))
