package scanner

import (
	"context"
	"fmt"

	"github.com/xssnick/tonutils-go/ton"
)

func (s *scanner) getShardID(shard *ton.BlockIDExt) string {
	return fmt.Sprintf("%d|%d", shard.Workchain, shard.Shard)
}

func (s *scanner) getNonSeenShards(
	ctx context.Context,
	shard *ton.BlockIDExt,
) (ret []*ton.BlockIDExt, err error) {
	if seqno, ok := s.shardLastSeqno[s.getShardID(shard)]; ok && seqno == shard.SeqNo {
		return nil, nil
	}

	block, err := s.api.GetBlockData(ctx, shard)
	if err != nil {
		return nil, fmt.Errorf("get block data err: ", err)
	}

	parents, err := block.BlockInfo.GetParentBlocks()
	if err != nil {
		return nil, fmt.Errorf("get parent blocks (%d:%d): %w", shard.Workchain, shard.Shard, err)
	}

	for _, parent := range parents {
		ext, err := s.getNonSeenShards(ctx, parent)
		if err != nil {
			return nil, err
		}

		ret = append(ret, ext...)
	}

	ret = append(ret, shard)
	return ret, nil
}

func (s *scanner) addBlock(master ton.BlockIDExt) error {
	s.lastBlock.SeqNo += 1
	return nil
}

func (s *scanner) getLastBlockSeqno() (uint32, error) {
	lastMaster, err := s.api.GetMasterchainInfo(context.Background())
	if err != nil {
		return 0, err
	}

	return lastMaster.SeqNo, nil
}
