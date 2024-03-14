import grpc

from diagonal_b6 import expression
from diagonal_b6 import VERSION

from diagonal_b6 import api_pb2
from diagonal_b6 import api_pb2_grpc

class Connection:

    def __init__(self, stub, root):
        self.stub = stub
        self.root = root

    def __call__(self, e):
        request = api_pb2.EvaluateRequestProto()
        request.version = VERSION
        node = expression.to_node(e)
        request.request.CopyFrom(node.to_node_proto())
        if self.root:
            request.root.CopyFrom(self.root.to_proto())
        return expression.from_node_proto(self.stub.Evaluate(request).result)

def connect_insecure(address, root=None, channel_arguments=None):
    channel = grpc.insecure_channel(address, options=channel_arguments)
    return Connection(api_pb2_grpc.B6Stub(channel), root)

def connect(address, token, root=None, root_certificates=None, channel_arguments=None):
    channel_credentials = grpc.ssl_channel_credentials(root_certificates=root_certificates)
    credentials = grpc.composite_channel_credentials(channel_credentials, grpc.access_token_call_credentials(token))
    channel = grpc.secure_channel(address, credentials, options=channel_arguments)
    return Connection(api_pb2_grpc.B6Stub(channel), root)
