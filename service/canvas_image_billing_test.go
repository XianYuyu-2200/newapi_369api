package service

import (
	"net/http/httptest"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCanvasFlatImageSettlementUsesHalfDollarPerImage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	price := hosttypes.PriceData{
		ModelPrice: 0.5,
		UsePrice:   true,
		GroupRatioInfo: hosttypes.GroupRatioInfo{
			GroupRatio: 1,
		},
	}
	price.AddOtherRatio("n", 2)
	info := &relaycommon.RelayInfo{
		OriginModelName: "expensive-image-model",
		PriceData:       price,
	}
	summary := calculateTextQuotaSummary(ctx, info, &dto.Usage{PromptTokens: 1, CompletionTokens: 0, TotalTokens: 1})
	require.Equal(t, 500000, summary.Quota)
	require.Nil(t, info.TieredBillingSnapshot)

	price = hosttypes.PriceData{
		ModelPrice: 0.5,
		UsePrice:   true,
		GroupRatioInfo: hosttypes.GroupRatioInfo{
			GroupRatio: 1,
		},
	}
	price.AddOtherRatio("n", 1)
	info.PriceData = price
	summary = calculateTextQuotaSummary(ctx, info, &dto.Usage{PromptTokens: 1, TotalTokens: 1})
	require.Equal(t, 250000, summary.Quota)
}
