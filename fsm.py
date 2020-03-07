
LogCommand  = 0

#  LogNoop is used to assert leadership.
LogNoop = 1

#  LogAddPeer is used to add a new peer. This should only be used with
#  older protocol versions designed to be compatible with unversioned
#  Raft servers. See comments in config.go for details.
LogAddPeerDeprecated =2

#  LogRemovePeer is used to remove an existing peer. This should only be
#  used with older protocol versions designed to be compatible with
#  unversioned Raft servers. See comments in config.go for details.
LogRemovePeerDeprecated =3 

#  LogBarrier is used to ensure all preceding operations have been
#  applied to the FSM. It is similar to LogNoop, but instead of returning
#  once committed, it only returns once the FSM manager acks it. Otherwise
#  it is possible there are operations committed but not yet applied to
#  the FSM.
LogBarrier =4 

#  LogConfiguration establishes a membership change configuration. It is
#  created when a server is added, removed, promoted, etc. Only used
#  when protocol version 1 or greater is in use.
LogConfiguration =5

class FSM:
    def apply(self, index, term, type, data):
        """ 
        Index holds the index of the log entry.
        Term holds the election term of the log entry.
        Type holds the type of the log entry.
        Data (bytes) holds the log entry's type-specific data.
        """
        pass

    def snapshot(self) -> bytes:
        """
        return 
        """
        pass

    def restore(bytes)
