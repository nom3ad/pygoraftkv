# import msgpackrpc
import gevent
from gevent.server import StreamServer
from gevent import socket, subprocess
from mprpc import RPCServer
import os
import os.path as path
import logging
import json
from collections import namedtuple

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

PYGORAFTKV_BIN =  path.join(path.abspath(path.dirname(__file__)), "pygoraftkv.bin")
m = {}
class BridgeServer(RPCServer):
    def __init__(self):
        RPCServer.__init__(self)

    def Set(self, k, v):
        print('Setting', k, '=', v)
        m[k] = v

    def Get(self, k):
        print("Get", k)
        return m.get(k)

    def Delete(self, k):
        print("Delete", k)
        return m.pop(k, None)

    def __call__(self, sock, addr):
        if not addr:
            addr = ('unix-local', -1)
        RPCServer.__call__(self, sock, addr)

Member = namedtuple("Member", ["id", "host", "port"])

def run_subprocess(bridge_addr, myid, quorum, raftdir):
    if not any(m.id == myid for m in quorum):
        raise ValueError("MyId %s is missing from quorum", myid)
    env = {
        'PGRKV_BRIDGE_ADDR': bridge_addr,
        'PGRKV_QUORUM' :  json.dumps([m._asdict() for m in quorum]),
        'PGRKV_MYID' : myid,
        'PGRKV_RAFTDIR' : raftdir
    }
   
    try:
        logger.debug("execing %s: env=%r", PYGORAFTKV_BIN, env)
        with subprocess.Popen(["pygorafty"], executable=PYGORAFTKV_BIN, shell=False,env=env) as sp:
            sp.wait()
            logger.info("Subprocess exited with code %d", sp.returncode)
            if sp.returncode != 0:
                raise Exception("Nonzero exit of subprocess")
    except:
        logger.error("Unexpected subprocess routine exit")
        raise

greenlets = []
def on_g_exit(g):
    logger.warning("Tearing down due to failure of %s", g)
    gevent.killall(greenlets)

def register(g):
    greenlets.append(g)
    g.link_exception(on_g_exit)

def run(myid, quorum, raftdir):
    listener = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) or ('127.0.0.1', 50000)
    sockname = './' + os.path.basename(__file__) + '.sock'
    if os.path.exists(sockname):
        os.remove(sockname)
    listener.bind(sockname)
    listener.listen(1)
    
    server = StreamServer(listener, BridgeServer())

    register(gevent.spawn(server.serve_forever))
    # register(gevent.spawn(run_subprocess, sockname, myid, quorum, raftdir))
    
    gevent.joinall(greenlets)


run('node1', [Member(id='node1',host="localhost", port=12345)],  '/tmp/node2')
