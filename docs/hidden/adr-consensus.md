# ADR 1: Consensus

## Requirements

- BFT up to 1/3 (1/3 is optimal)
- p2p gossip
- replicaiites an arbitrary state machine
- ~100 validators
- ~10000 txs per sec and ~1s latency
- accountability and deterministic replay
- partial synchrony

## Design

### Algorithm

Tendermint BFT is implemented as a state machine that changes state on receiving sets of votes from peers and timeouts from an internal clock.
The states of the internal consensus state machine are referred to as "steps".

The steps are:

- NewHeight - start consensus on a new block
- NewRound - start consensus on a new round
- Propose - we've proposed or started waiting for a proposal
- Prevote - we prevoted for a proposal or nil and we're waiting to hear +2/3 prevotes 
- PrevoteWait - we heard +2/3 prevotes for any (but not for a single block), and we're waiting to hear from more peers
- Precommit - we precommitted a polka or nil and we're waiting to hear +2/3 precommits
- PrecommitWait - we heard +2/3 precommits for any (but not for a single block), and we're waiting to hear from more peers
- Commit - we heard +2/3 precommits for a single block and we wrote the block to the chain.

### Consensus State

#### Structure

- Message sources include other peers, ourselves, and timeouts
- Messages include proposals, block parts, and votes
- One routine orders all msgs (`receiveRoutine`) to enable deterministic playback
- One routine maintains a timer (a `TimeoutTicker`), and sends timeouts to the receiveRoutine.
- TimeoutTicker schedules timeouts for a given H/R/S if that H/R/S has not already passed
- PrivValidator for signing votes
- RoundState to hold all votes and internal consensus state 
- BlockStore and Mempool interfaces to store committed blocks and lock the mempool during block commits.
- sm.State for block execution (should be able to refactor to an interface with MakeBlock,VerifyBlock, ApplyBlock)
- Three input channels for three sources of input:
  - peerMsgQueue for input from the outside world
  - internalMsgQueue for input from ourselves
  - timeoutTicker for input from timers
- evsw for notifications on significant events
- write-ahead log ensures recovery and the avoidance of signing conflicting votes
- closure actions for testing
- enterXXX methods for each consensus step

#### Execution

- Start by loading the last commits
- Schedule the openning NewHeight timeout 
- Msg handling:
	- proposal: verify and set internally
	- block part: verify and add internally. 
		- we may be waiting for block parts in ProposeStep or CommitStep
		- if it's the last piece, we'll go to prevote or to commit the block.
	- vote: verify and add it internally
		- execute the appropriate enterXXX method if any
	- timeout: execute the appropriate enterXXX method
- Messages are written to WAL before being handled
- Note that using only one receive routine means enterXXX are never executed concurrently
- Some enterXXX functions result in producing internal messages
	- messages are loaded on a buffered channel
	- if the channel is full, they are sent in go-routines to avoid blocking the receiveRoutine.
- Timeout messages from the TimeoutTicker are sent in a go-routine (ie. a new go-routine for each timeout!)
	- this is so the timeout routine doesn't block, so its always available to schedule new timeouts
	- determinism comes after receipt by the receiveRoutine
- From commit step, finalize the commit:
	- save the block to the store
	- write the last WAL msg for this height
	- apply the block to the app
	- fire new block events
	- update consensus state and schedule next round

- a complete description of where and when the enterXXX methods are called is:

enterNewRound
- handleTimeout: timeout NewHeight
- handleTimeout: timeout PrecommitWait. next round
- addVote: have all precommits and SkipTimeout==true. next round
- addVote: 2/3 precommits for nil. next round
- addVote: 2/3 prevote any
- addVote: 2/3 precommit for block. 
- addVote: 2/3 precommit for any

enterPropose
- enterNewRound: always at end

enterPrevote
- handleTimeout: timeout propose
- enterPropose: if proposal is complete
- addProposalBlockPart: proposal is complete
- addVote: waiting for a POL and now proposal complete
- addVote: +2/3 prevotes any for future/current round

enterPrevoteWait
- addVote: +2/3 prevotes any

enterPrecommit
- handleTimeout: timeout prevote
- addVote: polka
- addVote: 2/3 precommit for block or nil.
- addVote: 2/3 precommit for any. next round

enterPrecommitWait
- addVote: 2/3 precommit for any

enterCommit
- addVote: 2/3 precommit for block. 

### Reactor

#### Structure

- Track the consensus state of each peer using the p2p.Peer's concurrent data store
- Receive method inputs message to the consensus state and/or updates peer states. 
- Most gossip is handled by background routines that run for each peer. 
	- Routines repeatedly check peer state and internal consensus to determine what messages to send 
	- Three routines for 
		- gossip data (block parts)
		- gossip votes
		- gossip hints about correct blocks in the face of double signing
- Broadcasts: on receiving NewRoundStep and Vote events from the consensus state, all peers are alerted 

#### Execution


- Receive
	- peer state: 
		- ApplyNewRoundStep, ApplyCommitStep, ApplyHasVote, ApplyProposalPOL, ApplyVoteSetBits
		- SetHasProposal, SetHasProposalBlockPart, SetHasVote
		- EnsureVoteBitArrays
	- consensus input:
		- Proposal, BlockPart, Vote
	- votes
		- SetPeerMaj23 and respond with VoteSetBitsMessage
- TODO: detail the routines

### Tests

#### State

#### Reactor

#### Replay
