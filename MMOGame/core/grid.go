package core

import (
	"fmt"
	"sync"
)

type Grid struct {
	Gid int
	MinX int
	MaxX int
	MinY int
	MaxY int

	PlayerIDs map[int]bool
	playerLock	sync.RWMutex
}

func NewGrid(gid, minX, maxX, miny, maxY int) *Grid {
	return &Grid{
		Gid: gid,
		MinX: minX,
		MaxX: maxX,
		MinY: miny,
		MaxY: maxY,
		PlayerIDs: make(map[int]bool),
	}
}


// 给格子添加一个玩家
func (g *Grid) Add(playerId int) {
	g.playerLock.Lock()
	defer g.playerLock.Unlock()

	g.PlayerIDs[playerId] = true
	
}

//从格子中删除一个玩家
func (g *Grid) Remove(playerId int) {
	g.playerLock.RLock()
	defer g.playerLock.RUnlock()

	delete(g.PlayerIDs, playerId)
}

// 得到当前格子中的所有玩家
func (g *Grid) GetPlayerIDs() (playerIDs []int)	{
	g.playerLock.RLock()
	defer g.playerLock.RUnlock()

	for k, _ := range g.PlayerIDs {
		playerIDs = append(playerIDs, k)
	}
	return
}


func (g *Grid) String() string {
	return fmt.Sprintf("Grid ==== id:%d, minX:%d, maxX:%d, minY:%d, maxY:%d, playerIDs:%v",
		g.Gid, g.MinX, g.MaxX, g.MinY, g.MaxY, g.PlayerIDs)
}
