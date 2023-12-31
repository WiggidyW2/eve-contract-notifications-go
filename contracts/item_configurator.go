package contracts

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/WiggidyW/eve-contract-notifications-go/hashcode"
	pb "github.com/WiggidyW/eve-contract-notifications-go/proto"
)

const (
	ADDRESS       string = "##ITEM_CONFIGURATOR_SERVER_ADDRESS##"
	REFRESH_TOKEN string = "##ITEM_CONFIGURATOR_REFRESH_TOKEN##"

	MAX_RETRIES int           = 3
	RETRY_DELAY time.Duration = 10 * time.Second
)

type ItemConfiguratorClient struct {
	client pb.ItemConfiguratorClient
}

func NewItemConfiguratorClient(
	ctx context.Context,
) (*ItemConfiguratorClient, error) {
	conn, err := grpc.DialContext(
		ctx,
		ADDRESS,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}
	return &ItemConfiguratorClient{pb.NewItemConfiguratorClient(conn)}, nil
}

func (c *ItemConfiguratorClient) Get(ctx context.Context) (*[]Contract, error) {
	var rep *pb.BuybackContractsRep
	var err error

	// Retry a few times if we fail to get the contracts.
	for i := 0; i < MAX_RETRIES; i++ {
		// Sleep for a bit before retrying.
		if i > 0 {
			time.Sleep(RETRY_DELAY)
		}
		rep, err = c.client.BuybackContracts(ctx, Req())
		// If we succeeded, break out of the loop.
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get contracts: %v", err)
	}

	contracts := make([]Contract, 0, len(rep.Contracts))
	for _, c := range rep.Contracts {
		if c.EsiContract == nil {
			return nil, fmt.Errorf(
				"contract has no ESI contract: %v",
				c,
			)
		}
		contracts = append(contracts, Contract{
			HashCode: hashcode.HashCode(c.HashCode),
			Issued:   time.Unix(int64(c.EsiContract.Issued), 0),
			Expires:  time.Unix(int64(c.EsiContract.Expires), 0),
		})
	}

	return &contracts, nil
}

func Req() *pb.BuybackContractsReq {
	return &pb.BuybackContractsReq{
		IncludeItems: false,
		IncludeCheck: false,
		IncludeBuy:   false,
		RefreshToken: REFRESH_TOKEN,
		Language:     "en",
	}
}
