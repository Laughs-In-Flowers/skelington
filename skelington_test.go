package skelington

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

//func TestInfo(t *testing.T) {
//s, err := Default("stdout", "stack", , "bge")
//s, err := Default("stdout", "stack", , "rsp")
//spew.Dump(err)

//h, err := s.Process(100, "")
//if err != nil {
//	spew.Dump(err)
//	return
//}

//}

func testAllocator(expected int, file, allocator string, t *testing.T) {
	s, err := Default("null", "test", file, allocator)
	if err != nil {
		t.Errorf("error obtaining instance%s: %s", allocator, err)
	}
	h, err := s.Process("")
	if err != nil {
		t.Errorf("error processing instance %s: %s", allocator, err)
	}
	spew.Dump(h)
	spew.Dump(h.Statistics())
}

func TestRSPAllocator(t *testing.T) {
	testAllocator(97, "resources/skeleton_rsp.yaml", "rsp", t)
}

func TestBGEAllocator(t *testing.T) {
	//testAllocator(165, "resources/skeleton_bge.yaml", "bge", t)
}
