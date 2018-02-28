package skelington

import (
	"strings"
	"testing"
)

func TestSkelington(t *testing.T) {
	d, err := newProcessor(
		SkeletonRoot("test"),
	)
	if err != nil {
		t.Errorf("error obtaining test base: %s", err)
	}
	testPather(t, d.root, testPatherExpect)
	testSequence(t)
	testEMPAllocator(t)
	testRSPAllocator(t)
	testBGEAllocator(t)
	testEDFAllocator(t)
}

type patherExpect struct {
	key, path string
	tag       *Tag
}

var testPatherExpect *patherExpect = &patherExpect{
	"root", "test", &Tag{0, "test"},
}

func testPather(t *testing.T, p Pather, pe *patherExpect) {
	if p.Key() != pe.key {
		t.Errorf("testing pather key error: expected %s, got %s", pe.key, p.Key())
	}
	if p.Path() != pe.path {
		t.Errorf("testing pather path error: expected %s, got %s", pe.path, p.Path())
	}
	tg := p.Tag()
	if tg.Order != pe.tag.Order {
		t.Errorf("testing pather tag order error: expected %d, got %d", pe.tag.Order, tg.Order)
	}
	if tg.Value != pe.tag.Value {
		t.Errorf("testing pather tag value error: expected %s, got %s", pe.tag.Value, tg.Value)
	}
	p.SetPath("NEW_PATH")
	if p.Path() != "NEW_PATH" {
		t.Error("error setting new path on pather")
	}
}

func testSequence(t *testing.T) {
	ts := &Sequence{0, 0}
	if ts.String() != "0-of-0" {
		t.Error("sequence test error, expected '0-of-0'")
	}
}

func testHandle(t *testing.T, s *Skeleton) {
	hs := make(map[string]string)
	SkeletonCallsHandle(s,
		func(hh Handle) error {
			k := hh.Key()
			v := hh.Tagged(true)
			vv := strings.Join(v.List(), "")
			hs[k] = vv
			return nil
		},
		func(hh Handle) error {
			path1 := hh.Path()
			hh.SetPath("cannot set path")
			path2 := hh.Path()
			if path1 != path2 {
				t.Errorf("handle paths before and after being set shouild be equal: %s != %s", path1, path2)
			}
			return nil
		},
		func(hh Handle) error {
			ht := hh.Tag()
			if ht.Order != -1 {
				t.Errorf("error with handle tag: %v", ht)
			}
			return nil
		},
	)
	err := s.RunHook(HPost)
	if err != nil {
		t.Errorf("handler error: %s", err)
	}
	for _, v := range s.Has {
		id := v.Key()
		detail := v.Tagged(true)
		details := strings.Join(detail.List(), "")
		has, exists := hs[id]
		if !exists {
			t.Errorf("set map of compared handles is missing an expected key & value")
		}
		if has != details {
			t.Errorf("handle does not match the expected values set: %s != %s", has, details)
		}
	}
}

func compareStats(t *testing.T, have, expect map[string]int) {
	for k, v := range expect {
		var value int
		var exists bool
		if value, exists = have[k]; !exists {
			t.Errorf("%s not found, but expected", k)
		}
		if value != v {
			t.Errorf("Have %d but expected %d for key %s", value, v, k)
		}
	}
}

func testSkeleton(expected map[string]int,
	root, file, allocator, offset string,
	t *testing.T,
	hf ...HandleFunc) {
	s, err := New(
		SkeletonRoot(root),
		SkeletonFile(file),
		SkeletonAllocator(allocator),
		SkeletonAllocationOffset(offset),
	)
	SkeletonCallsStatistics(s, []StatFunc{}, []StatFunc{})
	if err != nil {
		t.Errorf("error with skeleton instance %s: %s", allocator, err)
	}
	testHandle(t, s)
	compareStats(t, s.Stat, expected)
}

func testEMPAllocator(t *testing.T) {
	exp := map[string]int{}
	testSkeleton(exp, "test", "", "", "", t)
}

func testRSPAllocator(t *testing.T) {
	exp := map[string]int{
		"KILL":         2,
		"BABYCARRIAGE": 1,
		"TOTAL":        97,
		"OTHER":        6,
		"SLOW":         2,
		"FAST":         6,
		"STRAIGHT":     4,
		"CURVED":       4,
		"OILSLICK":     6,
		"MISSLE":       6,
		"STOP":         2,
		"PILE":         13,
		"COW":          13,
		"TELEPORT":     6,
		"STAR":         6,
		"CAR":          7,
		"HOLE":         13,
	}

	exp0 := map[string]int{
		"BABYCARRIAGE": 1,
		"TOTAL":        40,
		"PILE":         13,
		"COW":          13,
		"HOLE":         13,
	}

	testSkeleton(exp, "test", "resources/skeleton_rsp.yaml", "rsp", "", t)
	testSkeleton(exp0, "test", "resources/skeleton_rsp.yaml", "rsp", "Obstacle", t)
}

func testBGEAllocator(t *testing.T) {
	exp := map[string]int{
		"DUST":         200,
		"STAR":         40,
		"TOTAL":        3302,
		"TYPE4":        20,
		"GALAXY":       20,
		"ROCKS":        1200,
		"TREES":        1200,
		"TYPE1":        100,
		"TYPE2":        40,
		"CIVILIZATION": 120,
		"TYPE3":        40,
		"PLANET":       120,
		"ASTEROID":     200,
		"UNIVERSE":     2,
	}

	exp0 := map[string]int{
		"TOTAL":        67,
		"ROCKS":        30,
		"TREES":        30,
		"CIVILIZATION": 3,
		"PLANET":       3,
		"STAR":         1,
	}

	testSkeleton(exp, "test", "resources/skeleton_bge.yaml", "bge", "", t)
	testSkeleton(exp0, "test", "resources/skeleton_bge.yaml", "bge", "Star", t)
}

func testEDFAllocator(t *testing.T) {
	exp := map[string]int{}

	testSkeleton(exp, "", "", "edf", "", t)
}
