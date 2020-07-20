package test

import (
	"common"
	"shared_svr/svr_rank/rank"
	"testing"
)

// go test -v ./src/svr_client/test/rank_test.go

type RankInfo struct {
	Score float64
	Pid   uint32
	Name  string
}

var g = rank.TScore{"season", "season_"}

func TestRank1(t *testing.T) {
	p1 := RankInfo{101, 41, "test1"}
	g.Set(p1.Name, p1.Score, &p1)

	p2 := RankInfo{102, 42, "test2"}
	g.Set(p2.Name, p2.Score, &p2)

	p3 := RankInfo{103, 43, "test3"}
	g.Set(p3.Name, p3.Score, &p3)

	if g.Rank("test1") != 2 || g.Rank("test2") != 1 || g.Rank("test3") != 0 {
		t.Fatal("Rank")
	}

	if b := g.GetInfo("test2"); b != nil {
		v := RankInfo{}
		if common.B2T(b, &v); v != p2 {
			t.Fatal("GetInfo")
		}
	}
	if g.GetScore("test2") != p2.Score {
		t.Fatal("GetScore")
	}
	if g.AddScore("test2", 10) != 112 {
		t.Fatal("AddScore")
	}

	if g.Rank("test1") != 2 || g.Rank("test2") != 0 || g.Rank("test3") != 1 {
		t.Fatal("Rank")
	}

	if g.Del("test2"); g.GetScore("test2") != 0 || g.GetInfo("test2") != nil {
		t.Fatal("Del")
	}
	if g.Clear(); g.GetInfo("test1") != nil {
		t.Fatal("Clear")
	}
}
