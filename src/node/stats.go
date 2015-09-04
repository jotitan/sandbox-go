package node

import "hardwareutil"


// Stats store data about node
type Stats struct {
	Memory int
	CPU int
	NbTaskers int
	NbTasks int
	Load float64
	Temperature float32
	ID string
}

func getStats(tm TasksManager)Stats{
	stats := Stats{}
	stats.CPU = hardwareutil.GetCPUUsage()
	stats.Memory = hardwareutil.GetCurrentMemory()
	stats.NbTaskers = tm.NbParallelTask
	stats.Load = tm.GetLoad()
	stats.NbTasks = len(tm.tasks)
	stats.Temperature = hardwareutil.GetTemperature()
	stats.ID = tm.url
	return stats
}
