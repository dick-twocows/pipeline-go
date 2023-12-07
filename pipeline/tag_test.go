package pipeline

import (
	"testing"
	"time"
)

func TestTag1(t *testing.T) {
	p := NewPipeline()
	defer p.Cancel()

	slice := EmptySlice[Tag[int]](p)

	// peek := Peek[Tag[int]](p, slice, f)

	remove := TagRemove[int](p, slice)

	<-ForEachStdOutV[int](p, remove).Output()
}

func TestTag2(t *testing.T) {
	p := NewPipeline()
	defer p.Cancel()

	slice := Slice[Tag[int]](p, []Tag[int]{&tag[int]{2, 2}, &tag[int]{3, 3}, &tag[int]{5, 5}, &tag[int]{9, 9}, &tag[int]{1, 1}, &tag[int]{7, 7}, &tag[int]{4, 4}, &tag[int]{8, 8}, &tag[int]{6, 6}})

	// limit := Limit[int](p, sequence, 10)

	// add := TagAdd[int](p, limit)

	f := func(p Pipeline, t Tag[int]) error {
		// fmt.Printf("Value [%v]\n", t)
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// mapper := Mapper[Tag[int]](p, add, f, MapperOpts{})

	peek := Peek[Tag[int]](p, slice, f)

	remove := TagRemove[int](p, peek)

	<-ForEachStdOutV[int](p, remove).Output()
}

func TestTag3(t *testing.T) {
	p := NewPipeline()
	defer p.Cancel()

	slice := Slice[Tag[int]](p, []Tag[int]{&tag[int]{2, 2}, &tag[int]{3, 3}, &tag[int]{5, 5}, &tag[int]{9, 9}, &tag[int]{1, 1}, &tag[int]{7, 7}, &tag[int]{4, 4}, &tag[int]{8, 8}})

	// limit := Limit[int](p, sequence, 10)

	// add := TagAdd[int](p, limit)

	f := func(p Pipeline, t Tag[int]) error {
		// fmt.Printf("Value [%v]\n", t)
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// mapper := Mapper[Tag[int]](p, add, f, MapperOpts{})

	peek := Peek[Tag[int]](p, slice, f)

	remove := TagRemove[int](p, peek)

	<-ForEachStdOutV[int](p, remove).Output()

	time.Sleep(2 * time.Second)
}
