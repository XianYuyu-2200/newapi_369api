package model

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAggregateGroupStatusUsesRealRequests(t *testing.T) {
	now := int64(1_800_000)
	start, end := GroupStatusBounds(now)
	require.Equal(t, GroupStatusBucketCount*GroupStatusBucketSeconds, end-start)
	require.LessOrEqual(t, now, end)
	require.GreaterOrEqual(t, now, end-GroupStatusBucketSeconds)

	t.Run("no requests stay empty", func(t *testing.T) {
		report := AggregateGroupStatus([]GroupStatusInput{{
			Name:        "vip",
			Description: "VIP group",
			Ratio:       0.5,
			Models:      []string{"gpt-4o", "gpt-4o-mini"},
		}}, nil, now)
		require.Len(t, report.Groups, 1)
		group := report.Groups[0]
		require.Equal(t, 2, group.ModelCount)
		require.True(t, group.Enabled)
		require.Equal(t, 0, group.Requests)
		require.Nil(t, group.Availability)
		require.Nil(t, group.AvgUseTime)
		require.Nil(t, group.AvgFirstToken)
		require.Nil(t, group.CacheRate)
		require.Nil(t, group.LatestAt)
		require.Equal(t, 2, group.ModelsWithoutRequests)
		require.Empty(t, group.Models)
		require.Len(t, group.History, GroupStatusBucketCount)
		for _, bucket := range group.History {
			require.Zero(t, bucket.Success)
			require.Zero(t, bucket.Error)
		}
	})

	t.Run("counts success failure latency and cache", func(t *testing.T) {
		logs := []GroupStatusLog{
			{
				Group: "vip", Type: LogTypeConsume, ModelName: "gpt-4o",
				UseTime: 30, PromptTokens: 100, CreatedAt: start,
				Other: `{"cache_tokens":40,"frt":18170}`,
			},
			{
				Group: "vip", Type: LogTypeConsume, ModelName: "gpt-4o",
				UseTime: 34, PromptTokens: 100, CreatedAt: start + 10,
				Other: `{"cache_tokens":10,"frt":10000}`,
			},
			{
				Group: "vip", Type: LogTypeConsume, ModelName: "gpt-4o-mini",
				UseTime: 2, PromptTokens: 50, CreatedAt: now,
				Other: `{"frt":500,"stream_status":{"status":"error"}}`,
			},
			{
				Group: "vip", Type: LogTypeError, ModelName: "gpt-4o",
				UseTime: 1, PromptTokens: 20, CreatedAt: now + 1,
				Other: `{"frt":9000}`,
			},
			{
				Group: "hidden", Type: LogTypeConsume, ModelName: "secret",
				UseTime: 9, PromptTokens: 100, CreatedAt: now,
				Other: `{"cache_tokens":100,"frt":1000}`,
			},
			{
				Group: "vip", Type: LogTypeConsume, ModelName: "old",
				UseTime: 99, PromptTokens: 100, CreatedAt: start - 1,
				Other: `{"cache_tokens":100,"frt":1000}`,
			},
		}
		report := AggregateGroupStatus([]GroupStatusInput{
			{Name: "svip", Description: "svip", Ratio: 2, Models: nil},
			{Name: "vip", Description: "VIP group", Ratio: 0.5, Models: []string{"gpt-4o", "gpt-4o-mini", "idle"}},
			{Name: "auto", Description: "auto", Ratio: 1},
		}, logs, now)
		require.Equal(t, []string{"vip", "svip"}, []string{report.Groups[0].Name, report.Groups[1].Name})

		group := report.Groups[0]
		require.Equal(t, 4, group.Requests)
		require.Equal(t, 2, group.Success)
		require.Equal(t, 2, group.Error)
		require.InDelta(t, 0.5, *group.Availability, 1e-9)
		require.InDelta(t, 32, *group.AvgUseTime, 1e-9)
		require.InDelta(t, 14.085, *group.AvgFirstToken, 1e-9)
		require.InDelta(t, 50.0/250.0, *group.CacheRate, 1e-9)
		require.Equal(t, now+1, *group.LatestAt)
		require.Equal(t, 1, group.ModelsWithoutRequests)
		require.Equal(t, 2, group.History[0].Success)
		require.Equal(t, 0, group.History[0].Error)
		require.Equal(t, 0, group.History[GroupStatusBucketCount-1].Success)
		require.Equal(t, 2, group.History[GroupStatusBucketCount-1].Error)
		require.Len(t, group.Models, 2)
		require.Equal(t, "gpt-4o", group.Models[0].Model)
		require.Equal(t, 3, group.Models[0].Requests)
		require.Equal(t, 2, group.Models[0].Success)
		require.Equal(t, 1, group.Models[0].Error)
		require.Nil(t, group.Models[1].AvgUseTime)
		require.Nil(t, group.Models[1].AvgFirstToken)

		require.Nil(t, report.Groups[1].Availability)
		require.False(t, report.Groups[1].Enabled)
	})
}

func TestParseGroupStatusOtherIgnoresMalformedMetadata(t *testing.T) {
	cacheTokens, frtSeconds, hasFrt, streamFailed := parseGroupStatusOther("{")
	require.Zero(t, cacheTokens)
	require.Zero(t, frtSeconds)
	require.False(t, hasFrt)
	require.False(t, streamFailed)

	cacheTokens, frtSeconds, hasFrt, streamFailed = parseGroupStatusOther(`{"cache_tokens":1.2,"frt":0,"stream_status":{"status":"ok"}}`)
	require.Equal(t, int64(1), cacheTokens)
	require.False(t, hasFrt)
	require.False(t, streamFailed)
	require.Zero(t, math.Abs(frtSeconds))
}
