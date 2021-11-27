package db

import (
	"testing"
	"context"
	"time"

	"prisma/gogroup"
	"prisma/tms/client_api"
	"prisma/tms"

	"go.uber.org/goleak"
	"github.com/stretchr/testify/require"
)

func TestTracksFeatureProvider_Service(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctx := gogroup.New(context.Background(), "test_service_stop")

	chSend := make(chan client_api.TrackUpdate)
	chReceive := make(chan FeatureUpdate)
	// check to finish releasing context
	f := TracksFeatureProvider{
		tracks: chSend,
		ProviderCommon: ProviderCommon{
			ctxt: ctx,
		},
	}
	go func() {
		require.EqualError(t, f.Service(chReceive), "context canceled")
	}()
	select {
		case chSend <- client_api.TrackUpdate{Track: &tms.Track{}}:
		case <- time.After(100 * time.Millisecond):
			t.Fatal("timeout")
	}
	select {
		case featUpd := <-chReceive:
			require.Nil(t, featUpd.Feature)
		case <- time.After(100 * time.Millisecond):
			t.Fatal("timeout")
	}
	ctx.Cancel(nil)
}
