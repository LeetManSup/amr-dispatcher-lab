package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"testing"
)

func TestObservedWaitIncludesQueuedTasks(t *testing.T) {
	task := &domain.TaskOrder{CreatedAt: 2, StartedAt: -1, FinishedAt: -1}
	if got := observedWait(task, 7); got != 5 {
		t.Fatalf("expected queued wait 5, got %d", got)
	}
}

func TestObservedWaitUsesActualStartWhenTaskStarted(t *testing.T) {
	task := &domain.TaskOrder{CreatedAt: 2, StartedAt: 6, FinishedAt: -1}
	if got := observedWait(task, 9); got != 4 {
		t.Fatalf("expected started wait 4, got %d", got)
	}
}

func TestObservedWaitUsesFinishWhenNeverStartedButClosed(t *testing.T) {
	task := &domain.TaskOrder{CreatedAt: 1, StartedAt: -1, FinishedAt: 5}
	if got := observedWait(task, 9); got != 4 {
		t.Fatalf("expected closed wait 4, got %d", got)
	}
}
