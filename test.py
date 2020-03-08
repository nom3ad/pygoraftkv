import sys

sys.path.append('./dist')
sys.path.append('.')


import dist.pygoraftkv as pg
import dist.go as go


quorum=pg.Slice_pygoraftkv_Member([pg.Member(Host="localhost", Port=1234, ID="1")])

raftdir="/tmp/data"
myId = "1"


class Bridge(go.GoClass):
    def __init__(self, *args, **kwargs):
        self.misc = 2
    
    def Set(self, afs, ival, sval):
        tfs = funcs.FunStruct(handle=afs)
        print("in python class fun: FieldI: ", tfs.FieldI, " FieldS: ", tfs.FieldS, " ival: ", ival, " sval: ", sval)

    def Get(self):
        fs.CallBack(77, self.ClassFun)

bridge = pg.FSMBridge(Bridge)



kv = pg.New(quorum, myId, raftdir, True, bridge)

import time
resp = kv.Open()
print('resp', resp)
# print('resp.error()', resp.Error()) # -- not working

print(kv.Get("abcd"))



i = 0
while True:
    time.sleep(2)
    i+=1
    print("LASt set was" , kv.Get("abcd"))
    print("setting...", i)
    try:
        print("Fut", kv.Set("abcd", str(i)))
    except RuntimeError as e:
        print(e)
