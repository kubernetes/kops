package paxos

type Participant interface {
	GossipState() GossipState
	Update(from GossipState) bool
	SetQuorum(uint)
	Propose()
	Think() bool
	Consensus() (bool, AcceptedValue)
	IsElector() bool
}

type Observer struct {
}

func NewObserver() Participant {
	return &Observer{}
}

func (observer *Observer) GossipState() GossipState {
	return nil
}

func (observer *Observer) Update(from GossipState) bool {
	return false
}

func (observer *Observer) Propose() {
}

func (observer *Observer) SetQuorum(uint) {
}

func (observer *Observer) Think() bool {
	return false
}

func (observer *Observer) Consensus() (bool, AcceptedValue) {
	return false, AcceptedValue{}
}

func (observer *Observer) IsElector() bool {
	return false
}
