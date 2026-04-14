package simulation

import (
	"amr-dispatcher-lab/internal/domain"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
)

func (e *Engine) robotUtilization(ticks float64) float64 {
	if len(e.robots) == 0 {
		return 0
	}
	totalBusyTicks := 0
	for _, robot := range e.robots {
		totalBusyTicks += robot.BusyTicks
	}
	return float64(totalBusyTicks) / (ticks * float64(len(e.robots)))
}

func (e *Engine) systemRisk(snapshot domain.SystemSnapshot) float64 {
	maxActive := e.admission.Config.MaxActiveTasks
	if maxActive <= 0 {
		maxActive = maxInt(1, len(e.robots))
	}
	return math.Min(1, float64(snapshot.ActiveTasks)/float64(maxActive))
}

func chooseStartTick(existing, tick int) int {
	if existing >= 0 {
		return existing
	}
	return tick
}

func computeUrgency(tick, deadline int) float64 {
	if deadline <= 0 {
		return 0
	}
	if deadline <= tick {
		return 1
	}
	return math.Min(1, 1/float64(deadline-tick))
}

func normalizePriority(priority int) float64 {
	if priority <= 0 {
		return 0
	}
	if priority > 10 {
		priority = 10
	}
	return float64(priority) / 10
}

func averagePathLoad(path []string, segmentLoad map[string]float64, segmentResolver func(string, string) string) float64 {
	if len(path) < 2 {
		return 0
	}
	total := 0.0
	count := 0.0
	for index := 0; index < len(path)-1; index++ {
		if segmentID := segmentResolver(path[index], path[index+1]); segmentID != "" {
			total += segmentLoad[segmentID]
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	mean := sum / float64(len(values))
	sumSquares := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}

func segmentRollingValues(window []map[string]float64) []float64 {
	values := make([]float64, 0)
	for _, snapshot := range window {
		for _, load := range snapshot {
			values = append(values, load)
		}
	}
	return values
}

func mergeAdmission(cfg, scenario domain.AdmissionConfig) domain.AdmissionConfig {
	result := scenario
	if cfg.MaxActiveTasks != 0 {
		result.MaxActiveTasks = cfg.MaxActiveTasks
	}
	if cfg.MaxRouteLoad != 0 {
		result.MaxRouteLoad = cfg.MaxRouteLoad
	}
	if cfg.MaxZoneLoad != 0 {
		result.MaxZoneLoad = cfg.MaxZoneLoad
	}
	if cfg.TaskTimeoutTicks != 0 {
		result.TaskTimeoutTicks = cfg.TaskTimeoutTicks
	}
	if cfg.StallThresholdTicks != 0 {
		result.StallThresholdTicks = cfg.StallThresholdTicks
	}
	return result
}

func mergeWeights(cfg, scenario domain.SchedulerWeights) domain.SchedulerWeights {
	result := scenario
	if cfg.Alpha != 0 {
		result.Alpha = cfg.Alpha
	}
	if cfg.Beta != 0 {
		result.Beta = cfg.Beta
	}
	if cfg.W1 != 0 {
		result.W1 = cfg.W1
	}
	if cfg.W2 != 0 {
		result.W2 = cfg.W2
	}
	if cfg.W3 != 0 {
		result.W3 = cfg.W3
	}
	if cfg.W4 != 0 {
		result.W4 = cfg.W4
	}
	if cfg.W5 != 0 {
		result.W5 = cfg.W5
	}
	return result
}

func hashRunConfig(cfg domain.RunConfig, mapName, scenarioName string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d|%d", cfg.Algorithm, mapName, scenarioName, cfg.Seed, cfg.TickDurationMillis)))
	return hex.EncodeToString(sum[:])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
