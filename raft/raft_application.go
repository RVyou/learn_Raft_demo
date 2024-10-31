package raft

// applicationTicker 日志确认
func (rf *Raft) applicationTicker() {
	for {
		select {
		default:
			rf.m.Lock()
			rf.applyCond.Wait() //等待有确认信息
			entries := make([]LogEntry, 0)
			//获取全部需要合并
			for i := rf.lastApplied; i <= rf.commitIndex; i++ {
				entries = append(entries, rf.log[i])
			}
			rf.m.Unlock()
			//整理
			for i, entry := range entries {
				rf.applyCh <- ApplyMsg{
					CommandValid: entry.CommandValid,
					Command:      entry.Command,
					CommandIndex: rf.lastApplied + 1 + i,
				}
			}
			rf.m.Lock()
			Info("[raft applicationTicker]", rf.lastApplied+1, "  change", rf.lastApplied+len(entries))
			rf.lastApplied += len(entries)
			if err := rf.persist(); err != nil {
				Error("[persist]", err)
			}
			rf.m.Unlock()
		}
	}
}
