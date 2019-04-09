package skelington

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type patherExpect struct {
	key, path string
	tag       *Tag
}

func TestPather(t *testing.T) {
	d, err := newProcessor(
		SetRoot("test"),
	)
	if err != nil {
		t.Errorf("error obtaining pather test base: %s", err)
	}
	p := d.root
	pe := &patherExpect{
		"root", "test", &Tag{0, "test"},
	}
	if p.Key() != pe.key {
		t.Errorf("testing pather key error: expected %s, got %s", pe.key, p.Key())
	}
	if p.Path() != pe.path {
		t.Errorf("testing pather path error: expected %s, got %s", pe.path, p.Path())
	}
	tg := p.GetTag()
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

func TestSequence(t *testing.T) {
	ts := &Sequence{0, 0}
	if ts.String() != "0-of-0" {
		t.Error("sequence test error, expected '0-of-0'")
	}
}

var tmpDir string = "/tmp/skelington/test"

func setup(t *testing.T, dir, name, block string) string {
	os.Mkdir(dir, os.ModeDir|os.ModePerm)
	var fileName string
	if name != "" && block != "" {
		fileName = filepath.Join(dir, name)
		f, err := open(fileName)
		if err != nil {
			t.Errorf("setup allocator error: %s", err)
			return fileName
		}
		b := new(bytes.Buffer)
		b.WriteString(block)
		f.Write(b.Bytes())
	}
	return fileName
}

func cleanup(t *testing.T, dirs ...string) {
	for _, dir := range dirs {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Errorf("cleanup allocator error: %s", err)
		}
	}
}

func testHandle(t *testing.T, s *Skelington) {
	hs := make(map[string]string)
	SkelingtonHandleCalls(s,
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
				t.Errorf("handle paths before and after being set should be equal: %s != %s", path1, path2)
			}
			return nil
		},
		func(hh Handle) error {
			ht := hh.GetTag()
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

func compareStats(instance string, t *testing.T, have, expect map[string]int) {
	for k, v := range expect {
		var haveValue int
		var exists bool
		if haveValue, exists = have[k]; !exists {
			t.Errorf("instance: %s --- %s not found, but expected for have stats", instance, k)
		}
		if haveValue != v {
			t.Errorf("instance: %s --- Have %d but expected %d for key %s for have stats", instance, haveValue, v, k)
		}
	}
}

func testSkelington(expected map[string]int,
	root, file, allocator, offset string,
	testEDF bool,
	t *testing.T) {
	coreSkelingtonTest(expected, root, file, allocator, offset, testEDF, t)
	if testEDF {
		coreSkelingtonTest(expected, filepath.Join(tmpDir, root), "no", "edf", DefaultSequencePatternString, false, t)
	}
}

func coreSkelingtonTest(expected map[string]int,
	root, file, allocator, offset string,
	testEDF bool,
	t *testing.T) {
	s, err := New(
		SetRoot(root),
		SetFile(file),
		SetAllocator(allocator),
		SetAllocationOffset(offset),
	)
	if err != nil {
		t.Errorf("error with skeleton instance %s: %s", allocator, err)
	}
	have := s.Report()
	haveC := make(map[string]int) // copy or testHandle might distort totals
	for k, v := range have {
		haveC[k] = v
	}
	testHandle(t, s)
	compareStats(allocator, t, haveC, expected)
	if testEDF {
		fsErr := toFile(s)
		if fsErr != nil {
			t.Errorf("Error transfering to file system: %s", fsErr.Error())
		}
	}
}

func toFile(s *Skelington) error {
	for _, h := range s.Has {
		mkPath := filepath.Join(tmpDir, h.Path())
		fsErr := os.MkdirAll(mkPath, os.ModeDir|os.ModePerm)
		if fsErr != nil {
			return fsErr
		}
	}
	return nil
}

func TestEMPAllocator(t *testing.T) {
	exp := map[string]int{}
	testSkelington(exp, "test", "", "", "", false, t)
}

var rspYaml = `---
-- tag: RSP_TEST
number: 100
levels:
  - tag: Car
    relative: false
    number: 7
  - tag: Road
    relative: true
    number: 14
    levels:
      - tag: Straight
        relative: true
        number: 50
      - tag: Curved
        relative: true
        number: 50
  - tag: Obstacle
    relative: false
    number: 40
    levels:
      - tag: Hole
        relative: false
        number: 13
      - tag: Pile
        relative: false
        number: 13
      - tag: Cow
        relative: false
        number: 13
      - tag: BabyCarriage
        relative: false
        number: 1
  - tag: PowerUps
    relative: true
    number: 50
    levels:
      - tag: Fast
        relative: true
      - tag: Missle
        relative: true
      - tag: Star
        relative: true
      - tag: Laser
        relative: true
        levels:
            - tag: Kill
              relative: true
            - tag: Slow
              relative: true
            - tag: Stop
              relative: true
      - tag: OilSlick
        relative: true
      - tag: Teleport
        relative: true
      - tag: Other
        relative: true`

var rspExp = map[string]int{
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

func TestRSPAllocator(t *testing.T) {
	fileName := setup(t, tmpDir, "rsp.yaml", rspYaml)

	exp := rspExp

	exp0 := map[string]int{
		"BABYCARRIAGE": 1,
		"TOTAL":        40,
		"PILE":         13,
		"COW":          13,
		"HOLE":         13,
	}

	testSkelington(exp, "testRSP", fileName, "rsp", "", true, t)
	testSkelington(exp0, "testRSPOffset", fileName, "rsp", "Obstacle", true, t)
	cleanup(t, tmpDir)
}

var bgeYaml = `---
-- tag: BGE_TEST
number: 0
levels:
  - tag: Universe
    leaf: true
    number: 2
    levels:
    - tag: Galaxy
      leaf: true
      number: 10
      levels:
        - tag: Star
          leaf: true
          number: 2
          levels:
          - tag: Planet
            leaf: true
            number: 3
            levels:
              - tag: Civilization
                number: 1
              - tag: Trees
                number: 10
              - tag: Rocks
                number: 10
        - tag: Asteroid
          number: 10
        - tag: BlackHole
          levels:
            - tag: Type1
              number: 5
            - tag: Type2
              number: 2
            - tag: Type3
              number: 2
            - tag: Type4
              number: 1
        - tag: Dust
          number: 10`

func TestBGEAllocator(t *testing.T) {
	fileName := setup(t, tmpDir, "bge.yaml", bgeYaml)

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

	testSkelington(exp, "testBGE", fileName, "bge", "", true, t)
	testSkelington(exp0, "testBGEOffset", fileName, "bge", "Star", true, t)
	cleanup(t, tmpDir)
}

/*
var currStats map[string]int

func skelingtonCallsStatistics(s *Skelington, once []StatFunc, every []StatFunc) {
	cs := newStat(s)
	for _, h := range s.Has {
		t := h.Unit()
		tag := strings.ToUpper(t.Value)
		once = append(once, func(s *Skelington, m map[string]int) error {
			cs.d[tag] = cs.d[tag] + 1
			return nil
		})
	}
	cs.Set("once", once...)

	s.AddHook(
		HPost,
		func(s *Skelington) error {
			err := cs.Run(s, "once")
			currStats = cs.Report()
			return err
		})
}
*/
