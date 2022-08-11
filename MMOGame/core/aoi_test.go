package core

import (
	"fmt"
	"testing"
)

func TestNewAOIManager(t *testing.T) {
	aoiMgr := NewAOIManager(0, 200, 0, 250, 4, 5)
	fmt.Println(aoiMgr)
}
