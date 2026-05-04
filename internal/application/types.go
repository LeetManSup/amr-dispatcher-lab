package application

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"amr-dispatcher-lab/internal/domain"
	"amr-dispatcher-lab/internal/simulation"
)

type CreateRunInput struct {
	Algorithm    domain.Algorithm `json:"algorithm"`
	MapPath      string           `json:"mapPath"`
	ScenarioPath string           `json:"scenarioPath"`
	Seed         int64            `json:"seed"`
}

type Catalog struct {
	Maps          []string      `json:"maps"`
	Scenarios     []string      `json:"scenarios"`
	MapItems      []CatalogItem `json:"mapItems,omitempty"`
	ScenarioItems []CatalogItem `json:"scenarioItems,omitempty"`
}

type Options struct {
	DefaultMapPath      string
	DefaultScenarioPath string
	TickDurationMillis  int
}

type Service struct {
	options           Options
	assets            ExperimentAssets
	store             RunStore
	clock             domain.Clock
	logger            *slog.Logger
	mu                sync.RWMutex
	runs              map[string]*runtime
	activeRunID       string
	counter           uint64
	eventMu           sync.RWMutex
	subscribers       map[uint64]liveSubscription
	eventCounter      uint64
	subscriberCounter uint64
}

type runtime struct {
	mu     sync.Mutex
	engine *simulation.Engine
	cancel context.CancelFunc
}

func NewService(options Options, assets ExperimentAssets, runStore RunStore, clock domain.Clock, logger *slog.Logger) *Service {
	if clock == nil {
		clock = domain.RealClock{}
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Service{
		options:     options,
		assets:      assets,
		store:       runStore,
		clock:       clock,
		logger:      logger.With("component", "application"),
		runs:        make(map[string]*runtime),
		subscribers: make(map[uint64]liveSubscription),
	}
}

func (s *Service) nextRunID() string {
	index := atomic.AddUint64(&s.counter, 1)
	return fmt.Sprintf("run-%d-%d", s.clock.Now().UnixNano(), index)
}
