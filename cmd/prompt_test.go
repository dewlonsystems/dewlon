/*
Copyright © 2026 Dewlon Systems
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestReadLineReturnsTheLine(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("hello\n"))

	got, err := readLine(context.Background(), in)
	if err != nil || got != "hello\n" {
		t.Fatalf("readLine = %q, %v", got, err)
	}
}

func TestReadLineAcceptsFinalLineWithoutNewline(t *testing.T) {
	in := bufio.NewReader(strings.NewReader("last"))

	got, err := readLine(context.Background(), in)
	if err != nil || got != "last" {
		t.Fatalf("readLine = %q, %v", got, err)
	}
}

func TestReadLineTreatsEmptyEOFAsCancel(t *testing.T) {
	in := bufio.NewReader(strings.NewReader(""))

	_, err := readLine(context.Background(), in)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected a cancellation, got %v", err)
	}
}

func TestReadLineStopsWhenContextIsCancelled(t *testing.T) {
	// A pipe nobody writes to: the read would block forever.
	pr, pw := io.Pipe()

	// Closing the write end also lets the background read finish, so the
	// goroutine doesn't outlive the test.
	defer func() { _ = pw.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	done := make(chan error, 1)
	go func() {
		_, err := readLine(ctx, bufio.NewReader(pr))
		done <- err
	}()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected a cancellation, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("readLine kept waiting after the context was cancelled")
	}
}
