package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPipelineBackground(t *testing.T) {
	p := Background()

	if p.ctx == nil {
		t.FailNow()
	}

	if p.cancel != nil {
		t.FailNow()
	}

	if p.err != nil {
		t.FailNow()
	}
}

func TestPipelineCancel(t *testing.T) {
	p := WithCancel(context.Background())

	if p.ctx == nil {
		t.Fatal("CTX is nil")
	}

	if p.cancel == nil {
		t.Fatal("cancel is nil")
	}

	if p.err != nil {
		t.Fatal("err not nil")
	}

	p.Cancel()
}

func TestPipelineTimeout(t *testing.T) {
	p := WithTimeout(context.Background(), time.Now().Add(2*time.Second))

	if p.ctx == nil {
		t.Fatal("CTX is nil")
	}

	if p.cancel == nil {
		t.Fatal("cancel is nil")
	}

	if p.err != nil {
		t.Fatal("err not nil")
	}

	<-p.Done()

}

func TestPipelineCancelWithError(t *testing.T) {
	p := WithCancel(context.Background())

	if p.ctx == nil {
		t.Fatal("CTX is nil")
	}

	if p.cancel == nil {
		t.Fatal("cancel is nil")
	}

	if p.err != nil {
		t.Fatal("err not nil")
	}

	p.CancelWithError(errors.New("foo"))

	if p.err == nil {
		t.Fatal("err is nil")
	}

}
