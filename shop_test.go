package shop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	comffC "github.com/comfforts/comff-constants"
	api "github.com/comfforts/comff-shops/api/v1"
	"github.com/comfforts/logger"
)

const TEST_DIR = "data"

func TestShopsClient(t *testing.T) {
	logger := logger.NewTestAppLogger(TEST_DIR)

	for scenario, fn := range map[string]func(
		t *testing.T,
		gc Client,
	){
		"shop CRUD, succeeds": testShopCRUD,
	} {
		t.Run(scenario, func(t *testing.T) {
			gc, teardown := setup(t, logger)
			defer teardown()
			fn(t, gc)
		})
	}

}

func setup(t *testing.T, logger logger.AppLogger) (
	gc Client,
	teardown func(),
) {
	t.Helper()

	clientOpts := NewDefaultClientOption()
	clientOpts.Caller = "shop-client-test"

	gc, err := NewClient(logger, clientOpts)
	require.NoError(t, err)

	return gc, func() {
		t.Logf(" TestGeoClient ended, will close geo client")
		err := gc.Close()
		require.NoError(t, err)
	}
}

func testShopCRUD(t *testing.T, sc Client) {
	t.Helper()
	rqstr, storeId, name, org := "client-shop-create-test@gmail.com", 1, "Client Create Test", "client-create-test"
	csr := api.AddShopRequest{
		Name:        name,
		Org:         org,
		StoreId:     uint64(storeId),
		City:        comffC.PETALUMA,
		PostalCode:  comffC.P94952,
		Country:     comffC.US,
		RequestedBy: rqstr,
	}

	ctx := context.Background()
	ctx1, cancel1 := context.WithCancel(ctx)
	defer cancel1()

	aResp, err := sc.AddShop(ctx1, &csr)
	require.NoError(t, err)
	require.Equal(t, storeId, int(aResp.Shop.StoreId))

	gResp, err := sc.GetShop(ctx, &api.GetShopRequest{
		Id: aResp.Shop.Id,
	})
	require.NoError(t, err)
	require.Equal(t, aResp.Shop.Id, gResp.Shop.Id)

	dResp, err := sc.DeleteShop(ctx, &api.DeleteShopRequest{
		Id: gResp.Shop.Id,
	})
	require.NoError(t, err)
	require.Equal(t, true, dResp.Ok)
}
