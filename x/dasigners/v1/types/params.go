package types

func (p *Params) Validate() error {
	if p.EpochBlocks == 0 {
		return ErrInvalidEpochBlocks
	}
	return nil
}
