package client

type FailMode int

const (
	//Failover selects another server automaticaly
	Failover FailMode = iota

	//Failfast returns error immediately
	Failfast

	//Failtry use current client again
	Failtry

	//Failbackup select another server if the first server doesn't respond in specified time and use the fast response.
	Failbackup
)

type SelectMode int

const (
	//RandomSelect is selecting randomly
	RandomSelect SelectMode = iota

	//RoundRobin is selecting by round robin
	RoundRobin

	//WeightedRoundRobin is selecting by weighted round robin
	WeightedRoundRobin

	//WeightedICMP is selecting by weighted Ping time
	WeightedICMP

	//ConsistentHash is selecting by hashing
	ConsistentHash

	//Closest is selecting the closest server
	Closest

	SelectByUser = 1000
)
