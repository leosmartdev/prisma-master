package db

import (
	"context"
	"prisma/gogroup"
	"testing"

	"go.uber.org/goleak"
)

func TestMiscDataFeatureProvider_Service(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := gogroup.New(context.Background(), "test_service_stop")
	// check to finish releasing context
	f := MiscDataFeatureProvider{
		data: make(chan GoGetResponse),
		ProviderCommon: ProviderCommon{
			ctxt: ctx,
		},
		table: &TableInfo{
			Name: "test",
		},
	}
	go f.Service(nil)
	ctx.Cancel(nil)
}
