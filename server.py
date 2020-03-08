# import msgpackrpc
from gevent.server import StreamServer
from mprpc import RPCServer

m = {}
class BridgeServer(RPCServer):
    def Set(self, k, v):
        print('Setting', k, '=', v)
        m[k] = v

    def Get(self, k):
        print("Get", k)
        return m.get(k)

    def Delete(self, k):
        print("Delete", k)
        return m.pop(k, None)



server = StreamServer(('127.0.0.1', 50000), BridgeServer())
server.serve_forever()

# server = msgpackrpc.Server(BridgeServer())
# server.listen(msgpackrpc.Address("localhost", 50000))
# server.start()