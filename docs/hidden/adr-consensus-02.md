
New design:
- internal state machine for handleMsg
- external state machine for reactor
- remove the internalMsgQueue ? then with 1-val we'd have no break points








TODO: move the proxyAppConn into the State so consensus doesn't need it


type State interface {
	MakeBlock(types.Txs, *types.Commit) *types.Block
	ValidateBlock(types.Block)
	ApplyBlock(eventCache, cs.proxyAppConn, block, blockPartsHeader, cs.mempool) // TODO: remove mempool dependence; but TxsAvailable
}


### Chain State
  - uses of State:
	- chainID, checking
	- lastblockid, validators, apphash
	- Validate block
  - State interface:
	- MakeBlock(txs, commit) // ensure height
	- ValidateBlock(block)
	err := stateCopy.ApplyBlock(eventCache, cs.proxyAppConn, block, blockParts.Header(), cs.mempool)

TODO: ChainID in the config





### Reactor






# Tendermint

- store
  - blocks (canonical blockchain and associated commits)
  - txs (tx indexing)
  - vals (historical validator sets)

- 
