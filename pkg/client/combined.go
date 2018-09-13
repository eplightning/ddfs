package client

type CombinedClient struct {
	*BlockClient
	*IndexClient
}

func NewCombinedClient(block *BlockClient, index *IndexClient) *CombinedClient {
	return &CombinedClient{
		BlockClient: block,
		IndexClient: index,
	}
}
