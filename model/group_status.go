package model

import (
	"encoding/json"
	"math"
	"sort"

	"github.com/QuantumNous/new-api/common"
)

const (
	GroupStatusBucketCount   = 30
	GroupStatusBucketSeconds = int64(60)
)

// GroupStatusBounds is the current 30 one-minute buckets, ending in the
// current minute. end is exclusive.
func GroupStatusBounds(now int64) (start, end int64) {
	endMinute := now / GroupStatusBucketSeconds
	startMinute := endMinute - GroupStatusBucketCount + 1
	return startMinute * GroupStatusBucketSeconds, (endMinute + 1) * GroupStatusBucketSeconds
}

type GroupStatusInput struct {
	Name        string
	Description string
	Ratio       float64
	Models      []string
}

type GroupStatusLog struct {
	Group        string `gorm:"column:group"`
	Type         int    `gorm:"column:type"`
	ModelName    string `gorm:"column:model_name"`
	UseTime      int    `gorm:"column:use_time"`
	PromptTokens int    `gorm:"column:prompt_tokens"`
	Other        string `gorm:"column:other"`
	CreatedAt    int64  `gorm:"column:created_at"`
}

type GroupStatusBucket struct {
	Success int `json:"success"`
	Error   int `json:"error"`
}

type GroupStatusModel struct {
	Model         string   `json:"model"`
	Requests      int      `json:"requests"`
	Success       int      `json:"success"`
	Error         int      `json:"error"`
	Availability  *float64 `json:"availability"`
	AvgUseTime    *float64 `json:"avgUseTime"`
	AvgFirstToken *float64 `json:"avgFirstToken"`
	CacheRate     *float64 `json:"cacheRate"`
}

type GroupStatus struct {
	Name                  string              `json:"name"`
	Description           string              `json:"description"`
	Ratio                 float64             `json:"ratio"`
	ModelCount            int                 `json:"modelCount"`
	Enabled               bool                `json:"enabled"`
	Requests              int                 `json:"requests"`
	Success               int                 `json:"success"`
	Error                 int                 `json:"error"`
	Availability          *float64            `json:"availability"`
	AvgUseTime            *float64            `json:"avgUseTime"`
	AvgFirstToken         *float64            `json:"avgFirstToken"`
	CacheRate             *float64            `json:"cacheRate"`
	LatestAt              *int64              `json:"latestAt"`
	ModelsWithoutRequests int                 `json:"modelsWithoutRequests"`
	History               []GroupStatusBucket `json:"history"`
	Models                []GroupStatusModel  `json:"models"`
}

type GroupStatusReport struct {
	WindowSeconds int64         `json:"windowSeconds"`
	BucketSeconds int64         `json:"bucketSeconds"`
	GeneratedAt   int64         `json:"generatedAt"`
	Groups        []GroupStatus `json:"groups"`
}

func ListEnabledModelsByGroup() (map[string][]string, error) {
	var abilities []Ability
	if err := DB.Where("enabled = ?", true).Find(&abilities).Error; err != nil {
		return nil, err
	}
	seen := make(map[string]map[string]struct{})
	for _, ability := range abilities {
		if !ability.Enabled || ability.Group == "" || ability.Model == "" {
			continue
		}
		if seen[ability.Group] == nil {
			seen[ability.Group] = make(map[string]struct{})
		}
		seen[ability.Group][ability.Model] = struct{}{}
	}
	result := make(map[string][]string, len(seen))
	for group, models := range seen {
		names := make([]string, 0, len(models))
		for name := range models {
			names = append(names, name)
		}
		sort.Strings(names)
		result[group] = names
	}
	return result, nil
}

func ListGroupStatusLogs(start, end int64, groups []string) ([]GroupStatusLog, error) {
	rows := make([]GroupStatusLog, 0)
	if len(groups) == 0 || start >= end {
		return rows, nil
	}
	err := LOG_DB.Model(&Log{}).
		Select(logGroupCol+", type, model_name, use_time, prompt_tokens, other, created_at").
		Where("created_at >= ? AND created_at < ?", start, end).
		Where("type IN ?", []int{LogTypeConsume, LogTypeError}).
		Where(logGroupCol+" IN ?", groups).
		Find(&rows).Error
	return rows, err
}

func AggregateGroupStatus(inputs []GroupStatusInput, logs []GroupStatusLog, now int64) GroupStatusReport {
	start, _ := GroupStatusBounds(now)
	accumulators := make(map[string]*groupStatusAccumulator, len(inputs))
	order := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if input.Name == "" || input.Name == "auto" {
			continue
		}
		if _, exists := accumulators[input.Name]; exists {
			continue
		}
		accumulators[input.Name] = newGroupStatusAccumulator(input)
		order = append(order, input.Name)
	}
	for _, row := range logs {
		accumulator := accumulators[row.Group]
		if accumulator == nil {
			continue
		}
		accumulator.add(row, start)
	}

	groups := make([]GroupStatus, 0, len(order))
	for _, name := range order {
		groups = append(groups, accumulators[name].export())
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Ratio != groups[j].Ratio {
			return groups[i].Ratio < groups[j].Ratio
		}
		return groups[i].Name < groups[j].Name
	})
	return GroupStatusReport{
		WindowSeconds: GroupStatusBucketCount * GroupStatusBucketSeconds,
		BucketSeconds: GroupStatusBucketSeconds,
		GeneratedAt:   now,
		Groups:        groups,
	}
}

type groupStatusAccumulator struct {
	input        GroupStatusInput
	enabled      map[string]struct{}
	success      int
	failure      int
	useTimeSum   float64
	useTimeCount int
	frtSum       float64
	frtCount     int
	cacheTokens  int64
	promptTokens int64
	latest       int64
	hasLatest    bool
	history      []GroupStatusBucket
	models       map[string]*modelStatusAccumulator
}

type modelStatusAccumulator struct {
	success      int
	failure      int
	useTimeSum   float64
	useTimeCount int
	frtSum       float64
	frtCount     int
	cacheTokens  int64
	promptTokens int64
}

func newGroupStatusAccumulator(input GroupStatusInput) *groupStatusAccumulator {
	enabled := make(map[string]struct{}, len(input.Models))
	for _, name := range input.Models {
		if name == "" {
			continue
		}
		enabled[name] = struct{}{}
	}
	return &groupStatusAccumulator{
		input:   input,
		enabled: enabled,
		history: make([]GroupStatusBucket, GroupStatusBucketCount),
		models:  make(map[string]*modelStatusAccumulator),
	}
}

func (g *groupStatusAccumulator) add(row GroupStatusLog, start int64) {
	if row.CreatedAt < start {
		return
	}
	bucket := int((row.CreatedAt - start) / GroupStatusBucketSeconds)
	if bucket < 0 || bucket >= len(g.history) {
		return
	}
	cacheTokens, frtSeconds, hasFrt, streamFailed := parseGroupStatusOther(row.Other)
	failed := row.Type == LogTypeError || streamFailed
	modelRow := g.models[row.ModelName]
	if modelRow == nil {
		modelRow = &modelStatusAccumulator{}
		g.models[row.ModelName] = modelRow
	}
	if failed {
		g.failure++
		g.history[bucket].Error++
		modelRow.failure++
	} else {
		g.success++
		g.history[bucket].Success++
		modelRow.success++
	}
	if !g.hasLatest || row.CreatedAt > g.latest {
		g.latest = row.CreatedAt
		g.hasLatest = true
	}
	if !failed && row.Type == LogTypeConsume {
		if row.UseTime > 0 {
			g.useTimeSum += float64(row.UseTime)
			g.useTimeCount++
			modelRow.useTimeSum += float64(row.UseTime)
			modelRow.useTimeCount++
		}
		if hasFrt {
			g.frtSum += frtSeconds
			g.frtCount++
			modelRow.frtSum += frtSeconds
			modelRow.frtCount++
		}
	}
	if row.Type == LogTypeConsume && row.PromptTokens > 0 {
		g.promptTokens += int64(row.PromptTokens)
		g.cacheTokens += cacheTokens
		modelRow.promptTokens += int64(row.PromptTokens)
		modelRow.cacheTokens += cacheTokens
	}
}

func (g *groupStatusAccumulator) export() GroupStatus {
	models := make([]GroupStatusModel, 0)
	seenRequests := make(map[string]struct{})
	for name, row := range g.models {
		requests := row.success + row.failure
		if requests == 0 {
			continue
		}
		seenRequests[name] = struct{}{}
		models = append(models, GroupStatusModel{
			Model:         name,
			Requests:      requests,
			Success:       row.success,
			Error:         row.failure,
			Availability:  availabilityRatio(row.success, row.failure),
			AvgUseTime:    averageRatio(row.useTimeSum, row.useTimeCount),
			AvgFirstToken: averageRatio(row.frtSum, row.frtCount),
			CacheRate:     cacheRatio(row.cacheTokens, row.promptTokens),
		})
	}
	sort.Slice(models, func(i, j int) bool {
		if models[i].Requests != models[j].Requests {
			return models[i].Requests > models[j].Requests
		}
		return models[i].Model < models[j].Model
	})
	missing := 0
	for name := range g.enabled {
		if _, ok := seenRequests[name]; !ok {
			missing++
		}
	}
	var latest *int64
	if g.hasLatest {
		value := g.latest
		latest = &value
	}
	return GroupStatus{
		Name:                  g.input.Name,
		Description:           g.input.Description,
		Ratio:                 g.input.Ratio,
		ModelCount:            len(g.enabled),
		Enabled:               len(g.enabled) > 0,
		Requests:              g.success + g.failure,
		Success:               g.success,
		Error:                 g.failure,
		Availability:          availabilityRatio(g.success, g.failure),
		AvgUseTime:            averageRatio(g.useTimeSum, g.useTimeCount),
		AvgFirstToken:         averageRatio(g.frtSum, g.frtCount),
		CacheRate:             cacheRatio(g.cacheTokens, g.promptTokens),
		LatestAt:              latest,
		ModelsWithoutRequests: missing,
		History:               g.history,
		Models:                models,
	}
}

func availabilityRatio(success, failure int) *float64 {
	total := success + failure
	if total <= 0 {
		return nil
	}
	value := float64(success) / float64(total)
	return &value
}

func averageRatio(sum float64, count int) *float64 {
	if count <= 0 {
		return nil
	}
	value := sum / float64(count)
	return &value
}

func cacheRatio(cacheTokens, promptTokens int64) *float64 {
	if promptTokens <= 0 {
		return nil
	}
	value := float64(cacheTokens) / float64(promptTokens)
	return &value
}

func parseGroupStatusOther(raw string) (cacheTokens int64, frtSeconds float64, hasFrt bool, streamFailed bool) {
	if raw == "" {
		return
	}
	var payload map[string]any
	if err := common.Unmarshal([]byte(raw), &payload); err != nil {
		return
	}
	if value, ok := metricNumber(payload["cache_tokens"]); ok && value > 0 {
		cacheTokens = int64(math.Round(value))
	}
	if value, ok := metricNumber(payload["frt"]); ok && value > 0 {
		frtSeconds = value / 1000
		hasFrt = true
	}
	if stream, ok := payload["stream_status"].(map[string]any); ok {
		if status, _ := stream["status"].(string); status == "error" {
			streamFailed = true
		}
	}
	return
}

func metricNumber(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case float32:
		return float64(number), true
	case int:
		return float64(number), true
	case int64:
		return float64(number), true
	case json.Number:
		parsed, err := number.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}
