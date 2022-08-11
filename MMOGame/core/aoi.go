package core

import (
	"fmt"
)

type AOIManager struct {
	MinX int
	MaxX int
	MinY int
	MaxY int

	CntX int
	CntY int

	grids map[int]*Grid
}

func NewAOIManager(minX, maxX, minY, maxY, cntX, cntY int) *AOIManager {
	aoiMgr :=  &AOIManager{
		MinX: minX,
		MaxX: maxX,
		MinY: minY,
		MaxY: maxY,
		CntX: cntX,
		CntY: cntY,
		grids: make(map[int]*Grid),
	}

	for y := 0; y < cntY; y++ {
		for x := 0; x < cntX; x ++ {
			//计算格子ID， 根据x, y
			//id = y*cntX + x
			gid := y * cntX + x
			
			//为给定的格子id分配对象
			aoiMgr.grids[gid] = NewGrid(gid, aoiMgr.MinX + aoiMgr.gridWidth() * x,
				aoiMgr.MinX + aoiMgr.gridWidth() * (x + 1),
				aoiMgr.MinY + aoiMgr.gridLength() * y,
				aoiMgr.MinY + aoiMgr.gridLength() * (y + 1))
		}
	}
	return aoiMgr
}

func (m *AOIManager) gridWidth() int {
	return (m.MaxX - m.MinX) / m.CntX
}

func (m *AOIManager) gridLength() int {
	return (m.MaxY - m.MinY) / m.CntY
}

func (m *AOIManager) String() string {
	s := fmt.Sprintf("AOIManager:\n, minX:%d, maxX:%d,cntX:%d, minY:%d, maxY:%d, cntY:%d\n, Grids in AOImanger:\n",
		 m.MinX, m.MaxX,m.CntX, m.MinY, m.MaxY, m.CntY)
	for _, grid := range m.grids {
		s += fmt.Sprintln(grid)
	}

	return s
}


//根据当前格子id得到九宫格信息
func (m *AOIManager) GetSurroundGridsByGid(GId int) (grids []*Grid) {
	//判断gid是否存在
	if _, ok := m.grids[GId]; !ok {
		return
	}

	// 将当前gid加入到九宫格当中
	grids = append(grids, m.grids[GId])

	idx := GId % m.CntX
	if idx > 0 {
		grids = append(grids, m.grids[GId-1])
	}

	if idx < m.CntX - 1 {
		grids = append(grids, m.grids[GId+1])
	}

	gridsX := make([]int, 0, len(grids))
	for _, v := range grids {
		gridsX = append(gridsX, v.Gid)
	}

	for _, v := range gridsX {
		idy := v / m.CntX
		if idy > 0 {
			grids = append(grids, m.grids[v-m.CntX])
		}

		if idy < m.CntY - 1 {
			grids = append(grids, m.grids[v+m.CntX])
		}

	}
	return
}

//通过GID获取当前格子的全部playerID
func (m *AOIManager) GetPidsByGid(gID int) (playerIDs []int) {
	playerIDs = m.grids[gID].GetPlayerIDs()
	return
}

//移除一个格子中的PlayerID
func (m *AOIManager) RemovePidFromGrid(pID, gID int) {
	m.grids[gID].Remove(pID)
}


//添加一个PlayerID到一个格子中
func (m *AOIManager) AddPidToGrid(pID, gID int) {
	m.grids[gID].Add(pID)
}


//通过横纵坐标添加一个Player到一个格子中
func (m *AOIManager) AddToGridByPos(pID int, x, y float32) {
	gID := m.GetGIDByPos(x, y)
	grid := m.grids[gID]
	grid.Add(pID)
}

//通过横纵坐标把一个Player从对应的格子中删除
func (m *AOIManager) RemoveFromGridByPos(pID int, x, y float32) {
	gID := m.GetGIDByPos(x, y)
	grid := m.grids[gID]
	grid.Remove(pID)
}

//通过横纵坐标获取对应的格子ID
func (m *AOIManager) GetGIDByPos(x, y float32) int {
	gx := (int(x) - m.MinX) / m.gridWidth()
	gy := (int(x) - m.MinY) / m.gridLength()

	return gy * m.CntX + gx
}

//通过横纵坐标得到周边九宫格内的全部PlayerIDs
func (m *AOIManager) GetPIDsByPos(x, y float32) (playerIDs []int) {
	//根据横纵坐标得到当前坐标属于哪个格子ID
	gID := m.GetGIDByPos(x, y)

	//根据格子ID得到周边九宫格的信息
	grids := m.GetSurroundGridsByGid(gID)
	for _, v := range grids {
		playerIDs = append(playerIDs, v.GetPlayerIDs()...)
		fmt.Printf("===> grid ID : %d, pids : %v  ====", v.Gid, v.GetPlayerIDs())
	}

	return
}