import sys

sys.path.append('./dist')
sys.path.append('.')


import dist.pygoraftkv as pg


quorum=pg.Slice_pygoraftkv_Peer([pg.Peer(Host="localhost", Port=1234, ID="1")])

print(quorum[0])
raftdir="/tmp/data"
myId = "1"

kv = pg.New(quorum, myId, raftdir, True)

import time
# kv.Open()
while True:
    time.sleep(1)
    pass