package gogroup

import (
	"errors"
	"testing"
	"time"
)

func TestGroupCancel(t *testing.T) {
	g := New(nil, "")
	started := make(chan bool, 1)
	gotExit := false
	g.Go(func(g2 GoGroup) {
		started <- true
		<-g2.Done()
		gotExit = true
	})

	ret := <-started
	if ret != true {
		t.Errorf("Got bad start")
	}

	g.Cancel(nil)
	g.Wait()

	if !gotExit {
		t.Errorf("Managed routine didn't exit!")
	}
}

func TestGroupRestarts(t *testing.T) {
	g := New(nil, "")

	starts := 0
	g.Go(func(incr int) {
		starts += incr
	}, 5)

	restarts := 0
	t.Logf("Running GoRestart()")
	g.GoRestart(func() {
		restarts += 1
		time.Sleep(time.Duration(1) * time.Millisecond)
		panic("Panic!")
	})

	t.Logf("Sleeping for a bit to allow restarts to go")
	time.Sleep(time.Duration(100) * time.Millisecond)
	t.Logf("Canceling")
	g.Cancel(nil)
	t.Logf("Waiting for group to exit")
	g.Wait()

	if starts != 5 {
		t.Errorf("Got %v starts, wanted 5", starts)
	}

	if restarts <= 1 {
		t.Errorf("Got %v restarts, wanted > 1", restarts)
	}
}

func TestGroupErrors(t *testing.T) {
	g := New(nil, "")

	g.Go(func() {
		panic("Panic!")
	})

	g.Go(func() error {
		return errors.New("Error!")
	})

	g.Go(func() error {
		return nil
	})

	errors := g.Wait()
	if len(errors) != 2 {
		t.Errorf("Got errors: %v, expected two of them", errors)
	}
	t.Logf("Errors: %+v", errors)
}

func TestGroupErrCallback(t *testing.T) {
	g := New(nil, "")
	numErrors := 0
	g.ErrCallback(func(err error) {
		numErrors += 1
		t.Logf("Got error: %+v", err)
	})

	g.Go(func() {
		panic("Panic!")
	})

	g.Go(func() error {
		return errors.New("Error!")
	})

	g.Wait()

	if numErrors != 2 {
		t.Errorf("Got %v errors, wanted 2", numErrors)
	}
}
