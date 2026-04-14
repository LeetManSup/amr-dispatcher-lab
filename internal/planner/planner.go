package planner

import (
	"amr-dispatcher-lab/internal/domain"
	"container/heap"
	"errors"
	"fmt"
	"math"
)

type RoutePlanner interface {
	BuildPath(from, to string) (domain.PathResult, error)
}

type GraphPlanner struct {
	points   map[string]domain.MapPoint
	edges    map[string][]edge
	segments map[string]domain.MapSegment
}

type edge struct {
	To        string
	Cost      float64
	SegmentID string
}

func NewGraphPlanner(m domain.MapData) (*GraphPlanner, error) {
	g := &GraphPlanner{
		points:   make(map[string]domain.MapPoint, len(m.Points)),
		edges:    make(map[string][]edge),
		segments: make(map[string]domain.MapSegment, len(m.Segments)),
	}
	for _, point := range m.Points {
		if point.PointID == "" {
			return nil, errors.New("point without id")
		}
		g.points[point.PointID] = point
	}
	for _, segment := range m.Segments {
		if _, ok := g.points[segment.FromPoint]; !ok {
			return nil, fmt.Errorf("unknown fromPoint %s", segment.FromPoint)
		}
		if _, ok := g.points[segment.ToPoint]; !ok {
			return nil, fmt.Errorf("unknown toPoint %s", segment.ToPoint)
		}
		if segment.SpeedLimit <= 0 {
			return nil, fmt.Errorf("segment %s has non-positive speedLimit", segment.SegmentID)
		}
		if segment.RouteCostWeight <= 0 {
			return nil, fmt.Errorf("segment %s has non-positive routeCostWeight", segment.SegmentID)
		}
		if segment.Length <= 0 {
			return nil, fmt.Errorf("segment %s has non-positive length", segment.SegmentID)
		}
		g.segments[segment.SegmentID] = segment
		cost := segment.Cost()
		switch segment.Direction {
		case domain.SegmentDirectionBidirectional:
			g.edges[segment.FromPoint] = append(g.edges[segment.FromPoint], edge{To: segment.ToPoint, Cost: cost, SegmentID: segment.SegmentID})
			g.edges[segment.ToPoint] = append(g.edges[segment.ToPoint], edge{To: segment.FromPoint, Cost: cost, SegmentID: segment.SegmentID})
		case domain.SegmentDirectionForward:
			g.edges[segment.FromPoint] = append(g.edges[segment.FromPoint], edge{To: segment.ToPoint, Cost: cost, SegmentID: segment.SegmentID})
		case domain.SegmentDirectionReverse:
			g.edges[segment.ToPoint] = append(g.edges[segment.ToPoint], edge{To: segment.FromPoint, Cost: cost, SegmentID: segment.SegmentID})
		default:
			return nil, fmt.Errorf("segment %s has unsupported direction %s", segment.SegmentID, segment.Direction)
		}
	}
	return g, nil
}

func (g *GraphPlanner) BuildPath(from, to string) (domain.PathResult, error) {
	if from == to {
		return domain.PathResult{Points: []string{from}, Cost: 0, ETA: 0}, nil
	}
	if _, ok := g.points[from]; !ok {
		return domain.PathResult{}, fmt.Errorf("unknown point %s", from)
	}
	if _, ok := g.points[to]; !ok {
		return domain.PathResult{}, fmt.Errorf("unknown point %s", to)
	}

	dist := make(map[string]float64, len(g.points))
	prev := make(map[string]string, len(g.points))
	for id := range g.points {
		dist[id] = math.Inf(1)
	}
	dist[from] = 0

	pq := &priorityQueue{}
	heap.Push(pq, &item{point: from, priority: 0})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*item)
		if current.point == to {
			break
		}
		if current.priority > dist[current.point] {
			continue
		}
		for _, next := range g.edges[current.point] {
			candidate := dist[current.point] + next.Cost
			if candidate < dist[next.To] {
				dist[next.To] = candidate
				prev[next.To] = current.point
				heap.Push(pq, &item{point: next.To, priority: candidate})
			}
		}
	}

	if math.IsInf(dist[to], 1) {
		return domain.PathResult{}, fmt.Errorf("path %s -> %s is unreachable", from, to)
	}

	path := []string{to}
	for cursor := to; cursor != from; {
		cursor = prev[cursor]
		path = append([]string{cursor}, path...)
	}

	return domain.PathResult{
		Points: path,
		Cost:   dist[to],
		ETA:    int(math.Ceil(dist[to])),
	}, nil
}

type item struct {
	point    string
	priority float64
	index    int
}

type priorityQueue []*item

func (p priorityQueue) Len() int            { return len(p) }
func (p priorityQueue) Less(i, j int) bool  { return p[i].priority < p[j].priority }
func (p priorityQueue) Swap(i, j int)       { p[i], p[j] = p[j], p[i]; p[i].index = i; p[j].index = j }
func (p *priorityQueue) Push(x interface{}) { *p = append(*p, x.(*item)) }
func (p *priorityQueue) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	*p = old[:n-1]
	return item
}
