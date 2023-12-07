package pipeline

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type tag[T any] struct {
	index int
	value T
}

func (t *tag[T]) Index() int {
	return t.index
}

func (t *tag[T]) Value() T {
	return t.value
}

type Tag[T any] interface {
	Index() int
	Value() T
}

func TagAdd[T any](p Pipeline, input Source[T]) *source[Tag[T]] {
	output := NewSource[Tag[T]](p, "TagAdd")

	go func() {
		index := 0

		defer func() {
			close(output.Output())
			output.Metrics("Index", index)
		}()

		for {
			select {
			case t, ok := <-input.Output():
				if !ok {
					return
				}
				index++
				tag := &tag[T]{index, t}
				select {
				case output.Output() <- tag:
				case <-p.Done():
					return
				}
			case <-p.Done():
				return
			}
		}
	}()

	return output
}

type node[T any] struct {
	tag  Tag[T]
	next atomic.Pointer[node[T]]
}

func TagRemove[T any](p Pipeline, input Source[Tag[T]]) *source[T] {
	output := NewSource[T](p, "TagRemove")

	head := atomic.Pointer[node[T]]{}

	addIndex := atomic.Int64{}
	addIndex.Store(1)
	removeIndex := atomic.Int64{}
	removeIndex.Store(1)

	// updateIndex := make(chan int)

	walkNodes := func() {
		cN := head.Load()
		// output.Logger().Info("Walking nodes")
		fmt.Println("Begin walking nodes")
		for cN != nil {
			StdOutConsumer[Tag[T]](p, cN.tag)
			cN = cN.next.Load()
		}
		fmt.Println("End walking nodes")
	}

	wg := sync.WaitGroup{}
	// eon := node[T]{}

	removeNewHead := make(chan *node[T])

	// Add
	wg.Add(1)
	go func() {
		// l := output.Logger().WithGroup("Add")

		index := 1

		// pruneCount := 0

		defer func() {
			close(removeNewHead)

			output.Metrics("AddCount", "Index", index)

			wg.Done()

			fmt.Println("ADD: OK")
		}()

		for {
			select {
			case t, ok := <-input.Output():
				if !ok {
					return
				}

				currentAddIndex := addIndex.Load()
				currentRemoveIndex := removeIndex.Load()
				fmt.Println("ADD: checking index", currentAddIndex, currentRemoveIndex)
				if currentRemoveIndex > currentAddIndex {
					cH := head.Load()
					cN := cH
					for cN != nil {
						if cN.tag.Index() >= int(currentRemoveIndex) {
							break
						}
						fmt.Println("ADD: removing node", cN)
						cN = cN.next.Load()
					}
					if cN != cH {
						fmt.Println("ADD: new head", cN, cN != nil)
						head.Store(cN)
					}
					addIndex.Store(currentRemoveIndex)
				}

				fmt.Println("ADD: Tag received", "Tag", t)

				cH := head.Load()
				hUpdated := false

				if cH == nil {
					nH := node[T]{t, atomic.Pointer[node[T]]{}}
					head.Store(&nH)
					hUpdated = true
				} else {
					var pN *node[T] = nil
					cN := cH
					added := false
					for !added {
						if t.Index() < cN.tag.Index() {
							if pN == nil {
								nN := node[T]{t, atomic.Pointer[node[T]]{}}
								nN.next.Store(cN)
								head.Store(&nN)
								hUpdated = true
								added = true
							} else {
								nN := node[T]{t, atomic.Pointer[node[T]]{}}
								nN.next.Store(pN.next.Load())
								pN.next.Store(&nN)
								added = true
							}
						} else {
							pN = cN
							cN = cN.next.Load()
							if cN == nil {
								nN := node[T]{t, atomic.Pointer[node[T]]{}}
								pN.next.Store(&nN)
								added = true
							}
						}
					}
				}

				if hUpdated {
					fmt.Println("ADD: send new head to remove")
					select {
					case removeNewHead <- head.Load():
					case <-p.Done():
						return
					}
				}

			case <-p.Done():
				return
			}
		}
	}()

	// Remove
	wg.Add(1)
	go func() {
		index := 1

		defer func() {
			fmt.Println("REMOVE: closing")

			output.Metrics("Index", index)

			wg.Done()

			fmt.Println("REMOVE: OK")

		}()

		for {
			select {
			case head, ok := <-removeNewHead:
				if !ok {
					fmt.Println("REMOVE: close new head")
					return
				}

				fmt.Println("REMOVE: head", head)

				// cI := index
				fmt.Println("REMOVE: index", index)

				cN := head
				for cN != nil {
					if cN.tag.Index() > index {
						break
					}
					if cN.tag.Index() == index {
						// fmt.Println("REMOVE: Output tag value", cN.tag)
						select {
						case output.Output() <- cN.tag.Value():
							index++
						case <-p.Done():
							return
						}
					}
					cN = cN.next.Load()
				}
				fmt.Println("REMOVE: index", index)
				removeIndex.Store(int64(index))
			case <-p.Done():
				return
			}
		}
	}()

	// TODO: Do we want to close the output in the reaper?
	// defer close(output)
	// If we have an error we can p.CancelWithError(...)
	go func() {
		fmt.Println("REAPER: waiting")

		defer func() {

			close(output.output)

		}()

		wg.Wait()

		currentHead := head.Load()
		currentAddIndex := addIndex.Load()
		currentRemoveIndex := removeIndex.Load()
		fmt.Println("REAPER: sanity checking index", currentAddIndex, currentRemoveIndex, currentHead)
		if currentRemoveIndex > currentAddIndex {
			cH := head.Load()
			cN := cH
			for cN != nil {
				if cN.tag.Index() >= int(currentRemoveIndex) {
					break
				}
				fmt.Println("REAPER: removing node", cN)
				cN = cN.next.Load()
			}
			if cN != cH {
				fmt.Println("REAPER: new head", cN, cN != nil)
				head.Store(cN)
			}
			addIndex.Store(currentRemoveIndex)
		}
		currentHead = head.Load()
		currentAddIndex = addIndex.Load()
		fmt.Println("REAPER: sanity checking index", currentAddIndex, currentRemoveIndex, currentHead)

		if currentHead != nil {
			fmt.Println("REAPER: inconsistent index before source closed", currentHead.tag.Index())
		}

		fmt.Println("REAPER: walk nodes")
		walkNodes()

		fmt.Println("REAPER: OK")

	}()

	return output
}
