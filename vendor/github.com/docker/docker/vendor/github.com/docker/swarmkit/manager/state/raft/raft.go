package raft

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/pkg/idutil"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/docker/go-events"
	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/ca"
	"github.com/docker/swarmkit/log"
	"github.com/docker/swarmkit/manager/raftselector"
	"github.com/docker/swarmkit/manager/state/raft/membership"
	"github.com/docker/swarmkit/manager/state/raft/storage"
	"github.com/docker/swarmkit/manager/state/store"
	"github.com/docker/swarmkit/watch"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/clock"
	"github.com/pkg/errors"
)

var (
	// ErrNoRaftMember is thrown when the node is not yet part of a raft cluster
	ErrNoRaftMember = errors.New("raft: node is not yet part of a raft cluster")
	// ErrConfChangeRefused is returned when there is an issue with the configuration change
	ErrConfChangeRefused = errors.New("raft: propose configuration change refused")
	// ErrApplyNotSpecified is returned during the creation of a raft node when no apply method was provided
	ErrApplyNotSpecified = errors.New("raft: apply method was not specified")
	// ErrAppendEntry is thrown when the node fail to append an entry to the logs
	ErrAppendEntry = errors.New("raft: failed to append entry to logs")
	// ErrSetHardState is returned when the node fails to set the hard state
	ErrSetHardState = errors.New("raft: failed to set the hard state for log append entry")
	// ErrApplySnapshot is returned when the node fails to apply a snapshot
	ErrApplySnapshot = errors.New("raft: failed to apply snapshot on raft node")
	// ErrStopped is returned when an operation was submitted but the node was stopped in the meantime
	ErrStopped = errors.New("raft: failed to process the request: node is stopped")
	// ErrLostLeadership is returned when an operation was submitted but the node lost leader status before it became committed
	ErrLostLeadership = errors.New("raft: failed to process the request: node lost leader status")
	// ErrRequestTooLarge is returned when a raft internal message is too large to be sent
	ErrRequestTooLarge = errors.New("raft: raft message is too large and can't be sent")
	// ErrCannotRemoveMember is thrown when we try to remove a member from the cluster but this would result in a loss of quorum
	ErrCannotRemoveMember = errors.New("raft: member cannot be removed, because removing it may result in loss of quorum")
	// ErrMemberRemoved is thrown when a node was removed from the cluster
	ErrMemberRemoved = errors.New("raft: member was removed from the cluster")
	// ErrNoClusterLeader is thrown when the cluster has no elected leader
	ErrNoClusterLeader = errors.New("raft: no elected cluster leader")
	// ErrMemberUnknown is sent in response to a message from an
	// unrecognized peer.
	ErrMemberUnknown = errors.New("raft: member unknown")
)

// LeadershipState indicates whether the node is a leader or follower.
type LeadershipState int

const (
	// IsLeader indicates that the node is a raft leader.
	IsLeader LeadershipState = iota
	// IsFollower indicates that the node is a raft follower.
	IsFollower
)

// EncryptionKeys are the current and, if necessary, pending DEKs with which to
// encrypt raft data
type EncryptionKeys struct {
	CurrentDEK []byte
	PendingDEK []byte
}

// EncryptionKeyRotator is an interface to find out if any keys need rotating.
type EncryptionKeyRotator interface {
	GetKeys() EncryptionKeys
	UpdateKeys(EncryptionKeys) error
	NeedsRotation() bool
	RotationNotify() chan struct{}
}

// Node represents the Raft Node useful
// configuration.
type Node struct {
	raftNode raft.Node
	cluster  *membership.Cluster

	raftStore           *raft.MemoryStorage
	memoryStore         *store.MemoryStore
	Config              *raft.Config
	opts                NodeOptions
	reqIDGen            *idutil.Generator
	wait                *wait
	campaignWhenAble    bool
	signalledLeadership uint32
	isMember            uint32

	// waitProp waits for all the proposals to be terminated before
	// shutting down the node.
	waitProp sync.WaitGroup

	confState     raftpb.ConfState
	appliedIndex  uint64
	snapshotIndex uint64
	writtenIndex  uint64

	ticker clock.Ticker
	doneCh chan struct{}
	// removeRaftCh notifies about node deletion from raft cluster
	removeRaftCh        chan struct{}
	removeRaftFunc      func()
	leadershipBroadcast *watch.Queue

	// used to coordinate shutdown
	// Lock should be used only in stop(), all other functions should use RLock.
	stopMu sync.RWMutex
	// used for membership management checks
	membershipLock sync.Mutex

	snapshotInProgress chan uint64
	asyncTasks         sync.WaitGroup

	// stopped chan is used for notifying grpc handlers that raft node going
	// to stop.
	stopped chan struct{}

	lastSendToMember    map[uint64]chan struct{}
	raftLogger          *storage.EncryptedRaftLogger
	keyRotator          EncryptionKeyRotator
	rotationQueued      bool
	waitForAppliedIndex uint64
}

// NodeOptions provides node-level options.
type NodeOptions struct {
	// ID is the node's ID, from its certificate's CN field.
	ID string
	// Addr is the address of this node's listener
	Addr string
	// ForceNewCluster defines if we have to force a new cluster
	// because we are recovering from a backup data directory.
	ForceNewCluster bool
	// JoinAddr is the cluster to join. May be an empty string to create
	// a standalone cluster.
	JoinAddr string
	// Config is the raft config.
	Config *raft.Config
	// StateDir is the directory to store durable state.
	StateDir string
	// TickInterval interval is the time interval between raft ticks.
	TickInterval time.Duration
	// ClockSource is a Clock interface to use as a time base.
	// Leave this nil except for tests that are designed not to run in real
	// time.
	ClockSource clock.Clock
	// SendTimeout is the timeout on the sending messages to other raft
	// nodes. Leave this as 0 to get the default value.
	SendTimeout    time.Duration
	TLSCredentials credentials.TransportCredentials

	KeyRotator EncryptionKeyRotator
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewNode generates a new Raft node
func NewNode(opts NodeOptions) *Node {
	cfg := opts.Config
	if cfg == nil {
		cfg = DefaultNodeConfig()
	}
	if opts.TickInterval == 0 {
		opts.TickInterval = time.Second
	}
	if opts.SendTimeout == 0 {
		opts.SendTimeout = 2 * time.Second
	}

	raftStore := raft.NewMemoryStorage()

	n := &Node{
		cluster:   membership.NewCluster(2 * cfg.ElectionTick),
		raftStore: raftStore,
		opts:      opts,
		Config: &raft.Config{
			ElectionTick:    cfg.ElectionTick,
			HeartbeatTick:   cfg.HeartbeatTick,
			Storage:         raftStore,
			MaxSizePerMsg:   cfg.MaxSizePerMsg,
			MaxInflightMsgs: cfg.MaxInflightMsgs,
			Logger:          cfg.Logger,
		},
		doneCh:              make(chan struct{}),
		removeRaftCh:        make(chan struct{}),
		stopped:             make(chan struct{}),
		leadershipBroadcast: watch.NewQueue(),
		lastSendToMember:    make(map[uint64]chan struct{}),
		keyRotator:          opts.KeyRotator,
	}
	n.memoryStore = store.NewMemoryStore(n)

	if opts.ClockSource == nil {
		n.ticker = clock.NewClock().NewTicker(opts.TickInterval)
	} else {
		n.ticker = opts.ClockSource.NewTicker(opts.TickInterval)
	}

	n.reqIDGen = idutil.NewGenerator(uint16(n.Config.ID), time.Now())
	n.wait = newWait()

	n.removeRaftFunc = func(n *Node) func() {
		var removeRaftOnce sync.Once
		return func() {
			removeRaftOnce.Do(func() {
				close(n.removeRaftCh)
			})
		}
	}(n)

	return n
}

// WithContext returns context which is cancelled when parent context cancelled
// or node is stopped.
func (n *Node) WithContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
		case <-n.stopped:
			cancel()
		}
	}()
	return ctx, cancel
}

// JoinAndStart joins and starts the raft server
func (n *Node) JoinAndStart(ctx context.Context) (err error) {
	ctx, cancel := n.WithContext(ctx)
	defer func() {
		cancel()
		if err != nil {
			n.done()
		}
	}()

	loadAndStartErr := n.loadAndStart(ctx, n.opts.ForceNewCluster)
	if loadAndStartErr != nil && loadAndStartErr != storage.ErrNoWAL {
		return loadAndStartErr
	}

	snapshot, err := n.raftStore.Snapshot()
	// Snapshot never returns an error
	if err != nil {
		panic("could not get snapshot of raft store")
	}

	n.confState = snapshot.Metadata.ConfState
	n.appliedIndex = snapshot.Metadata.Index
	n.snapshotIndex = snapshot.Metadata.Index
	n.writtenIndex, _ = n.raftStore.LastIndex() // lastIndex always returns nil as an error

	if loadAndStartErr == storage.ErrNoWAL {
		if n.opts.JoinAddr != "" {
			c, err := n.ConnectToMember(n.opts.JoinAddr, 10*time.Second)
			if err != nil {
				return err
			}
			client := api.NewRaftMembershipClient(c.Conn)
			defer func() {
				_ = c.Conn.Close()
			}()

			joinCtx, joinCancel := context.WithTimeout(ctx, 10*time.Second)
			defer joinCancel()
			resp, err := client.Join(joinCtx, &api.JoinRequest{
				Addr: n.opts.Addr,
			})
			if err != nil {
				return err
			}

			n.Config.ID = resp.RaftID

			if _, err := n.newRaftLogs(n.opts.ID); err != nil {
				return err
			}

			n.raftNode = raft.StartNode(n.Config, []raft.Peer{})

			if err := n.registerNodes(resp.Members); err != nil {
				n.raftLogger.Close(ctx)
				return err
			}
		} else {
			// First member in the cluster, self-assign ID
			n.Config.ID = uint64(rand.Int63()) + 1
			peer, err := n.newRaftLogs(n.opts.ID)
			if err != nil {
				return err
			}
			n.raftNode = raft.StartNode(n.Config, []raft.Peer{peer})
			n.campaignWhenAble = true
		}
		atomic.StoreUint32(&n.isMember, 1)
		return nil
	}

	if n.opts.JoinAddr != "" {
		log.G(ctx).Warning("ignoring request to join cluster, because raft state already exists")
	}
	n.campaignWhenAble = true
	n.raftNode = raft.RestartNode(n.Config)
	atomic.StoreUint32(&n.isMember, 1)
	return nil
}

// DefaultNodeConfig returns the default config for a
// raft node that can be modified and customized
func DefaultNodeConfig() *raft.Config {
	return &raft.Config{
		HeartbeatTick:   1,
		ElectionTick:    3,
		MaxSizePerMsg:   math.MaxUint16,
		MaxInflightMsgs: 256,
		Logger:          log.L,
		CheckQuorum:     true,
	}
}

// DefaultRaftConfig returns a default api.RaftConfig.
func DefaultRaftConfig() api.RaftConfig {
	return api.RaftConfig{
		KeepOldSnapshots:           0,
		SnapshotInterval:           10000,
		LogEntriesForSlowFollowers: 500,
		ElectionTick:               3,
		HeartbeatTick:              1,
	}
}

// MemoryStore returns the memory store that is kept in sync with the raft log.
func (n *Node) MemoryStore() *store.MemoryStore {
	return n.memoryStore
}

func (n *Node) done() {
	n.cluster.Clear()

	n.ticker.Stop()
	n.leadershipBroadcast.Close()
	n.cluster.PeersBroadcast.Close()
	n.memoryStore.Close()

	close(n.doneCh)
}

// Run is the main loop for a Raft node, it goes along the state machine,
// acting on the messages received from other Raft nodes in the cluster.
//
// Before running the main loop, it first starts the raft node based on saved
// cluster state. If no saved state exists, it starts a single-node cluster.
func (n *Node) Run(ctx context.Context) error {
	ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", n.Config.ID)))
	ctx, cancel := context.WithCancel(ctx)

	// nodeRemoved indicates that node was stopped due its removal.
	nodeRemoved := false

	defer func() {
		cancel()
		n.stop(ctx)
		if nodeRemoved {
			// Move WAL and snapshot out of the way, since
			// they are no longer usable.
			if err := n.raftLogger.Clear(ctx); err != nil {
				log.G(ctx).WithError(err).Error("failed to move wal after node removal")
			}
			// clear out the DEKs
			if err := n.keyRotator.UpdateKeys(EncryptionKeys{}); err != nil {
				log.G(ctx).WithError(err).Error("could not remove DEKs")
			}
		}
		n.done()
	}()

	wasLeader := false

	for {
		select {
		case <-n.ticker.C():
			n.raftNode.Tick()
			n.cluster.Tick()
		case rd := <-n.raftNode.Ready():
			raftConfig := n.getCurrentRaftConfig()

			// Save entries to storage
			if err := n.saveToStorage(ctx, &raftConfig, rd.HardState, rd.Entries, rd.Snapshot); err != nil {
				log.G(ctx).WithError(err).Error("failed to save entries to storage")
			}

			if len(rd.Messages) != 0 {
				// Send raft messages to peers
				if err := n.send(ctx, rd.Messages); err != nil {
					log.G(ctx).WithError(err).Error("failed to send message to members")
				}
			}

			// Apply snapshot to memory store. The snapshot
			// was applied to the raft store in
			// saveToStorage.
			if !raft.IsEmptySnap(rd.Snapshot) {
				// Load the snapshot data into the store
				if err := n.restoreFromSnapshot(rd.Snapshot.Data, false); err != nil {
					log.G(ctx).WithError(err).Error("failed to restore from snapshot")
				}
				n.appliedIndex = rd.Snapshot.Metadata.Index
				n.snapshotIndex = rd.Snapshot.Metadata.Index
				n.confState = rd.Snapshot.Metadata.ConfState
			}

			// If we cease to be the leader, we must cancel any
			// proposals that are currently waiting for a quorum to
			// acknowledge them. It is still possible for these to
			// become committed, but if that happens we will apply
			// them as any follower would.

			// It is important that we cancel these proposals before
			// calling processCommitted, so processCommitted does
			// not deadlock.

			if rd.SoftState != nil {
				if wasLeader && rd.SoftState.RaftState != raft.StateLeader {
					wasLeader = false
					if atomic.LoadUint32(&n.signalledLeadership) == 1 {
						atomic.StoreUint32(&n.signalledLeadership, 0)
						n.leadershipBroadcast.Publish(IsFollower)
					}

					// It is important that we set n.signalledLeadership to 0
					// before calling n.wait.cancelAll. When a new raft
					// request is registered, it checks n.signalledLeadership
					// afterwards, and cancels the registration if it is 0.
					// If cancelAll was called first, this call might run
					// before the new request registers, but
					// signalledLeadership would be set after the check.
					// Setting signalledLeadership before calling cancelAll
					// ensures that if a new request is registered during
					// this transition, it will either be cancelled by
					// cancelAll, or by its own check of signalledLeadership.
					n.wait.cancelAll()
				} else if !wasLeader && rd.SoftState.RaftState == raft.StateLeader {
					wasLeader = true
				}
			}

			// Process committed entries
			for _, entry := range rd.CommittedEntries {
				if err := n.processCommitted(ctx, entry); err != nil {
					log.G(ctx).WithError(err).Error("failed to process committed entries")
				}
			}

			// in case the previous attempt to update the key failed
			n.maybeMarkRotationFinished(ctx)

			// Trigger a snapshot every once in awhile
			if n.snapshotInProgress == nil &&
				(n.needsSnapshot() || raftConfig.SnapshotInterval > 0 &&
					n.appliedIndex-n.snapshotIndex >= raftConfig.SnapshotInterval) {
				n.doSnapshot(ctx, raftConfig)
			}

			if wasLeader && atomic.LoadUint32(&n.signalledLeadership) != 1 {
				// If all the entries in the log have become
				// committed, broadcast our leadership status.
				if n.caughtUp() {
					atomic.StoreUint32(&n.signalledLeadership, 1)
					n.leadershipBroadcast.Publish(IsLeader)
				}
			}

			// Advance the state machine
			n.raftNode.Advance()

			// On the first startup, or if we are the only
			// registered member after restoring from the state,
			// campaign to be the leader.
			if n.campaignWhenAble {
				members := n.cluster.Members()
				if len(members) >= 1 {
					n.campaignWhenAble = false
				}
				if len(members) == 1 && members[n.Config.ID] != nil {
					if err := n.raftNode.Campaign(ctx); err != nil {
						panic("raft: cannot campaign to be the leader on node restore")
					}
				}
			}

		case snapshotIndex := <-n.snapshotInProgress:
			if snapshotIndex > n.snapshotIndex {
				n.snapshotIndex = snapshotIndex
			}
			n.snapshotInProgress = nil
			n.maybeMarkRotationFinished(ctx)
			if n.rotationQueued && n.needsSnapshot() {
				// there was a key rotation that took place before while the snapshot
				// was in progress - we have to take another snapshot and encrypt with the new key
				n.rotationQueued = false
				n.doSnapshot(ctx, n.getCurrentRaftConfig())
			}
		case <-n.keyRotator.RotationNotify():
			// There are 2 separate checks:  rotationQueued, and n.needsSnapshot().
			// We set rotationQueued so that when we are notified of a rotation, we try to
			// do a snapshot as soon as possible.  However, if there is an error while doing
			// the snapshot, we don't want to hammer the node attempting to do snapshots over
			// and over.  So if doing a snapshot fails, wait until the next entry comes in to
			// try again.
			switch {
			case n.snapshotInProgress != nil:
				n.rotationQueued = true
			case n.needsSnapshot():
				n.doSnapshot(ctx, n.getCurrentRaftConfig())
			}
		case <-n.removeRaftCh:
			nodeRemoved = true
			// If the node was removed from other members,
			// send back an error to the caller to start
			// the shutdown process.
			return ErrMemberRemoved
		case <-ctx.Done():
			return nil
		}
	}
}

func (n *Node) needsSnapshot() bool {
	if n.waitForAppliedIndex == 0 && n.keyRotator.NeedsRotation() {
		keys := n.keyRotator.GetKeys()
		if keys.PendingDEK != nil {
			n.raftLogger.RotateEncryptionKey(keys.PendingDEK)
			// we want to wait for the last index written with the old DEK to be commited, else a snapshot taken
			// may have an index less than the index of a WAL written with an old DEK.  We want the next snapshot
			// written with the new key to supercede any WAL written with an old DEK.
			n.waitForAppliedIndex = n.writtenIndex
			// if there is already a snapshot at this index, bump the index up one, because we want the next snapshot
			if n.waitForAppliedIndex == n.snapshotIndex {
				n.waitForAppliedIndex++
			}
		}
	}
	return n.waitForAppliedIndex > 0 && n.waitForAppliedIndex <= n.appliedIndex
}

func (n *Node) maybeMarkRotationFinished(ctx context.Context) {
	if n.waitForAppliedIndex > 0 && n.waitForAppliedIndex <= n.snapshotIndex {
		// this means we tried to rotate - so finish the rotation
		if err := n.keyRotator.UpdateKeys(EncryptionKeys{CurrentDEK: n.raftLogger.EncryptionKey}); err != nil {
			log.G(ctx).WithError(err).Error("failed to update encryption keys after a successful rotation")
		} else {
			n.waitForAppliedIndex = 0
		}
	}
}

func (n *Node) getCurrentRaftConfig() api.RaftConfig {
	raftConfig := DefaultRaftConfig()
	n.memoryStore.View(func(readTx store.ReadTx) {
		clusters, err := store.FindClusters(readTx, store.ByName(store.DefaultClusterName))
		if err == nil && len(clusters) == 1 {
			raftConfig = clusters[0].Spec.Raft
		}
	})
	return raftConfig
}

// Done returns channel which is closed when raft node is fully stopped.
func (n *Node) Done() <-chan struct{} {
	return n.doneCh
}

func (n *Node) stop(ctx context.Context) {
	n.stopMu.Lock()
	defer n.stopMu.Unlock()

	close(n.stopped)

	n.waitProp.Wait()
	n.asyncTasks.Wait()

	n.raftNode.Stop()
	n.ticker.Stop()
	n.raftLogger.Close(ctx)
	atomic.StoreUint32(&n.isMember, 0)
	// TODO(stevvooe): Handle ctx.Done()
}

// isLeader checks if we are the leader or not, without the protection of lock
func (n *Node) isLeader() bool {
	if !n.IsMember() {
		return false
	}

	if n.Status().Lead == n.Config.ID {
		return true
	}
	return false
}

// IsLeader checks if we are the leader or not, with the protection of lock
func (n *Node) IsLeader() bool {
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	return n.isLeader()
}

// leader returns the id of the leader, without the protection of lock and
// membership check, so it's caller task.
func (n *Node) leader() uint64 {
	return n.Status().Lead
}

// Leader returns the id of the leader, with the protection of lock
func (n *Node) Leader() (uint64, error) {
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if !n.IsMember() {
		return raft.None, ErrNoRaftMember
	}
	leader := n.leader()
	if leader == raft.None {
		return raft.None, ErrNoClusterLeader
	}

	return leader, nil
}

// ReadyForProposals returns true if the node has broadcasted a message
// saying that it has become the leader. This means it is ready to accept
// proposals.
func (n *Node) ReadyForProposals() bool {
	return atomic.LoadUint32(&n.signalledLeadership) == 1
}

func (n *Node) caughtUp() bool {
	// obnoxious function that always returns a nil error
	lastIndex, _ := n.raftStore.LastIndex()
	return n.appliedIndex >= lastIndex
}

// Join asks to a member of the raft to propose
// a configuration change and add us as a member thus
// beginning the log replication process. This method
// is called from an aspiring member to an existing member
func (n *Node) Join(ctx context.Context, req *api.JoinRequest) (*api.JoinResponse, error) {
	nodeInfo, err := ca.RemoteNode(ctx)
	if err != nil {
		return nil, err
	}

	fields := logrus.Fields{
		"node.id": nodeInfo.NodeID,
		"method":  "(*Node).Join",
		"raft_id": fmt.Sprintf("%x", n.Config.ID),
	}
	if nodeInfo.ForwardedBy != nil {
		fields["forwarder.id"] = nodeInfo.ForwardedBy.NodeID
	}
	log := log.G(ctx).WithFields(fields)
	log.Debug("")

	// can't stop the raft node while an async RPC is in progress
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	n.membershipLock.Lock()
	defer n.membershipLock.Unlock()

	if !n.IsMember() {
		return nil, ErrNoRaftMember
	}

	if !n.isLeader() {
		return nil, ErrLostLeadership
	}

	// Find a unique ID for the joining member.
	var raftID uint64
	for {
		raftID = uint64(rand.Int63()) + 1
		if n.cluster.GetMember(raftID) == nil && !n.cluster.IsIDRemoved(raftID) {
			break
		}
	}

	remoteAddr := req.Addr

	// If the joining node sent an address like 0.0.0.0:4242, automatically
	// determine its actual address based on the GRPC connection. This
	// avoids the need for a prospective member to know its own address.

	requestHost, requestPort, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid address %s in raft join request", remoteAddr)
	}

	requestIP := net.ParseIP(requestHost)
	if requestIP != nil && requestIP.IsUnspecified() {
		remoteHost, _, err := net.SplitHostPort(nodeInfo.RemoteAddr)
		if err != nil {
			return nil, err
		}
		remoteAddr = net.JoinHostPort(remoteHost, requestPort)
	}

	// We do not bother submitting a configuration change for the
	// new member if we can't contact it back using its address
	if err := n.checkHealth(ctx, remoteAddr, 5*time.Second); err != nil {
		return nil, err
	}

	err = n.addMember(ctx, remoteAddr, raftID, nodeInfo.NodeID)
	if err != nil {
		log.WithError(err).Errorf("failed to add member %x", raftID)
		return nil, err
	}

	var nodes []*api.RaftMember
	for _, node := range n.cluster.Members() {
		nodes = append(nodes, &api.RaftMember{
			RaftID: node.RaftID,
			NodeID: node.NodeID,
			Addr:   node.Addr,
		})
	}
	log.Debugf("node joined")

	return &api.JoinResponse{Members: nodes, RaftID: raftID}, nil
}

// checkHealth tries to contact an aspiring member through its advertised address
// and checks if its raft server is running.
func (n *Node) checkHealth(ctx context.Context, addr string, timeout time.Duration) error {
	conn, err := n.ConnectToMember(addr, timeout)
	if err != nil {
		return err
	}

	if timeout != 0 {
		tctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = tctx
	}

	defer conn.Conn.Close()

	if err := conn.HealthCheck(ctx); err != nil {
		return errors.Wrap(err, "could not connect to prospective new cluster member using its advertised address")
	}

	return nil
}

// addMember submits a configuration change to add a new member on the raft cluster.
func (n *Node) addMember(ctx context.Context, addr string, raftID uint64, nodeID string) error {
	node := api.RaftMember{
		RaftID: raftID,
		NodeID: nodeID,
		Addr:   addr,
	}

	meta, err := node.Marshal()
	if err != nil {
		return err
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  raftID,
		Context: meta,
	}

	// Wait for a raft round to process the configuration change
	return n.configure(ctx, cc)
}

// updateMember submits a configuration change to change a member's address.
func (n *Node) updateMember(ctx context.Context, addr string, raftID uint64, nodeID string) error {
	node := api.RaftMember{
		RaftID: raftID,
		NodeID: nodeID,
		Addr:   addr,
	}

	meta, err := node.Marshal()
	if err != nil {
		return err
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeUpdateNode,
		NodeID:  raftID,
		Context: meta,
	}

	// Wait for a raft round to process the configuration change
	return n.configure(ctx, cc)
}

// Leave asks to a member of the raft to remove
// us from the raft cluster. This method is called
// from a member who is willing to leave its raft
// membership to an active member of the raft
func (n *Node) Leave(ctx context.Context, req *api.LeaveRequest) (*api.LeaveResponse, error) {
	nodeInfo, err := ca.RemoteNode(ctx)
	if err != nil {
		return nil, err
	}

	ctx, cancel := n.WithContext(ctx)
	defer cancel()

	fields := logrus.Fields{
		"node.id": nodeInfo.NodeID,
		"method":  "(*Node).Leave",
		"raft_id": fmt.Sprintf("%x", n.Config.ID),
	}
	if nodeInfo.ForwardedBy != nil {
		fields["forwarder.id"] = nodeInfo.ForwardedBy.NodeID
	}
	log.G(ctx).WithFields(fields).Debug("")

	if err := n.removeMember(ctx, req.Node.RaftID); err != nil {
		return nil, err
	}

	return &api.LeaveResponse{}, nil
}

// CanRemoveMember checks if a member can be removed from
// the context of the current node.
func (n *Node) CanRemoveMember(id uint64) bool {
	return n.cluster.CanRemoveMember(n.Config.ID, id)
}

func (n *Node) removeMember(ctx context.Context, id uint64) error {
	// can't stop the raft node while an async RPC is in progress
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if !n.IsMember() {
		return ErrNoRaftMember
	}

	if !n.isLeader() {
		return ErrLostLeadership
	}

	n.membershipLock.Lock()
	defer n.membershipLock.Unlock()
	if n.cluster.CanRemoveMember(n.Config.ID, id) {
		cc := raftpb.ConfChange{
			ID:      id,
			Type:    raftpb.ConfChangeRemoveNode,
			NodeID:  id,
			Context: []byte(""),
		}
		err := n.configure(ctx, cc)
		return err
	}

	return ErrCannotRemoveMember
}

// RemoveMember submits a configuration change to remove a member from the raft cluster
// after checking if the operation would not result in a loss of quorum.
func (n *Node) RemoveMember(ctx context.Context, id uint64) error {
	ctx, cancel := n.WithContext(ctx)
	defer cancel()
	return n.removeMember(ctx, id)
}

// processRaftMessageLogger is used to lazily create a logger for
// ProcessRaftMessage. Usually nothing will be logged, so it is useful to avoid
// formatting strings and allocating a logger when it won't be used.
func (n *Node) processRaftMessageLogger(ctx context.Context, msg *api.ProcessRaftMessageRequest) *logrus.Entry {
	fields := logrus.Fields{
		"method": "(*Node).ProcessRaftMessage",
	}

	if n.IsMember() {
		fields["raft_id"] = fmt.Sprintf("%x", n.Config.ID)
	}

	if msg != nil && msg.Message != nil {
		fields["from"] = fmt.Sprintf("%x", msg.Message.From)
	}

	return log.G(ctx).WithFields(fields)
}

// ProcessRaftMessage calls 'Step' which advances the
// raft state machine with the provided message on the
// receiving node
func (n *Node) ProcessRaftMessage(ctx context.Context, msg *api.ProcessRaftMessageRequest) (*api.ProcessRaftMessageResponse, error) {
	if msg == nil || msg.Message == nil {
		n.processRaftMessageLogger(ctx, msg).Debug("received empty message")
		return &api.ProcessRaftMessageResponse{}, nil
	}

	// Don't process the message if this comes from
	// a node in the remove set
	if n.cluster.IsIDRemoved(msg.Message.From) {
		n.processRaftMessageLogger(ctx, msg).Debug("received message from removed member")
		return nil, grpc.Errorf(codes.NotFound, "%s", ErrMemberRemoved.Error())
	}

	var sourceHost string
	peer, ok := peer.FromContext(ctx)
	if ok {
		sourceHost, _, _ = net.SplitHostPort(peer.Addr.String())
	}

	n.cluster.ReportActive(msg.Message.From, sourceHost)

	ctx, cancel := n.WithContext(ctx)
	defer cancel()

	// Reject vote requests from unreachable peers
	if msg.Message.Type == raftpb.MsgVote {
		member := n.cluster.GetMember(msg.Message.From)
		if member == nil || member.Conn == nil {
			n.processRaftMessageLogger(ctx, msg).Debug("received message from unknown member")
			return &api.ProcessRaftMessageResponse{}, nil
		}

		healthCtx, cancel := context.WithTimeout(ctx, time.Duration(n.Config.ElectionTick)*n.opts.TickInterval)
		defer cancel()

		if err := member.HealthCheck(healthCtx); err != nil {
			n.processRaftMessageLogger(ctx, msg).Debug("member which sent vote request failed health check")
			return &api.ProcessRaftMessageResponse{}, nil
		}
	}

	if msg.Message.Type == raftpb.MsgProp {
		// We don't accepted forwarded proposals. Our
		// current architecture depends on only the leader
		// making proposals, so in-flight proposals can be
		// guaranteed not to conflict.
		n.processRaftMessageLogger(ctx, msg).Debug("dropped forwarded proposal")
		return &api.ProcessRaftMessageResponse{}, nil
	}

	// can't stop the raft node while an async RPC is in progress
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if n.IsMember() {
		if err := n.raftNode.Step(ctx, *msg.Message); err != nil {
			n.processRaftMessageLogger(ctx, msg).WithError(err).Debug("raft Step failed")
		}
	}

	return &api.ProcessRaftMessageResponse{}, nil
}

// ResolveAddress returns the address reaching for a given node ID.
func (n *Node) ResolveAddress(ctx context.Context, msg *api.ResolveAddressRequest) (*api.ResolveAddressResponse, error) {
	if !n.IsMember() {
		return nil, ErrNoRaftMember
	}

	nodeInfo, err := ca.RemoteNode(ctx)
	if err != nil {
		return nil, err
	}

	fields := logrus.Fields{
		"node.id": nodeInfo.NodeID,
		"method":  "(*Node).ResolveAddress",
		"raft_id": fmt.Sprintf("%x", n.Config.ID),
	}
	if nodeInfo.ForwardedBy != nil {
		fields["forwarder.id"] = nodeInfo.ForwardedBy.NodeID
	}
	log.G(ctx).WithFields(fields).Debug("")

	member := n.cluster.GetMember(msg.RaftID)
	if member == nil {
		return nil, grpc.Errorf(codes.NotFound, "member %x not found", msg.RaftID)
	}
	return &api.ResolveAddressResponse{Addr: member.Addr}, nil
}

func (n *Node) getLeaderConn() (*grpc.ClientConn, error) {
	leader, err := n.Leader()
	if err != nil {
		return nil, err
	}

	if leader == n.Config.ID {
		return nil, raftselector.ErrIsLeader
	}
	l := n.cluster.GetMember(leader)
	if l == nil {
		return nil, errors.New("no leader found")
	}
	if !n.cluster.Active(leader) {
		return nil, errors.New("leader marked as inactive")
	}
	if l.Conn == nil {
		return nil, errors.New("no connection to leader in member list")
	}
	return l.Conn, nil
}

// LeaderConn returns current connection to cluster leader or raftselector.ErrIsLeader
// if current machine is leader.
func (n *Node) LeaderConn(ctx context.Context) (*grpc.ClientConn, error) {
	cc, err := n.getLeaderConn()
	if err == nil {
		return cc, nil
	}
	if err == raftselector.ErrIsLeader {
		return nil, err
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cc, err := n.getLeaderConn()
			if err == nil {
				return cc, nil
			}
			if err == raftselector.ErrIsLeader {
				return nil, err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// registerNode registers a new node on the cluster memberlist
func (n *Node) registerNode(node *api.RaftMember) error {
	if n.cluster.IsIDRemoved(node.RaftID) {
		return nil
	}

	member := &membership.Member{}

	existingMember := n.cluster.GetMember(node.RaftID)
	if existingMember != nil {
		// Member already exists

		// If the address is different from what we thought it was,
		// update it. This can happen if we just joined a cluster
		// and are adding ourself now with the remotely-reachable
		// address.
		if existingMember.Addr != node.Addr {
			member.RaftMember = node
			member.Conn = existingMember.Conn
			n.cluster.AddMember(member)
		}

		return nil
	}

	// Avoid opening a connection to the local node
	if node.RaftID != n.Config.ID {
		// We don't want to impose a timeout on the grpc connection. It
		// should keep retrying as long as necessary, in case the peer
		// is temporarily unavailable.
		var err error
		if member, err = n.ConnectToMember(node.Addr, 0); err != nil {
			return err
		}
	}

	member.RaftMember = node
	err := n.cluster.AddMember(member)
	if err != nil {
		if member.Conn != nil {
			_ = member.Conn.Close()
		}
		return err
	}

	return nil
}

// registerNodes registers a set of nodes in the cluster
func (n *Node) registerNodes(nodes []*api.RaftMember) error {
	for _, node := range nodes {
		if err := n.registerNode(node); err != nil {
			return err
		}
	}

	return nil
}

// ProposeValue calls Propose on the raft and waits
// on the commit log action before returning a result
func (n *Node) ProposeValue(ctx context.Context, storeAction []*api.StoreAction, cb func()) error {
	ctx, cancel := n.WithContext(ctx)
	defer cancel()
	_, err := n.processInternalRaftRequest(ctx, &api.InternalRaftRequest{Action: storeAction}, cb)
	if err != nil {
		return err
	}
	return nil
}

// GetVersion returns the sequence information for the current raft round.
func (n *Node) GetVersion() *api.Version {
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if !n.IsMember() {
		return nil
	}

	status := n.Status()
	return &api.Version{Index: status.Commit}
}

// SubscribePeers subscribes to peer updates in cluster. It sends always full
// list of peers.
func (n *Node) SubscribePeers() (q chan events.Event, cancel func()) {
	return n.cluster.PeersBroadcast.Watch()
}

// GetMemberlist returns the current list of raft members in the cluster.
func (n *Node) GetMemberlist() map[uint64]*api.RaftMember {
	memberlist := make(map[uint64]*api.RaftMember)
	members := n.cluster.Members()
	leaderID, err := n.Leader()
	if err != nil {
		leaderID = raft.None
	}

	for id, member := range members {
		reachability := api.RaftMemberStatus_REACHABLE
		leader := false

		if member.RaftID != n.Config.ID {
			if !n.cluster.Active(member.RaftID) {
				reachability = api.RaftMemberStatus_UNREACHABLE
			}
		}

		if member.RaftID == leaderID {
			leader = true
		}

		memberlist[id] = &api.RaftMember{
			RaftID: member.RaftID,
			NodeID: member.NodeID,
			Addr:   member.Addr,
			Status: api.RaftMemberStatus{
				Leader:       leader,
				Reachability: reachability,
			},
		}
	}

	return memberlist
}

// Status returns status of underlying etcd.Node.
func (n *Node) Status() raft.Status {
	return n.raftNode.Status()
}

// GetMemberByNodeID returns member information based
// on its generic Node ID.
func (n *Node) GetMemberByNodeID(nodeID string) *membership.Member {
	members := n.cluster.Members()
	for _, member := range members {
		if member.NodeID == nodeID {
			return member
		}
	}
	return nil
}

// IsMember checks if the raft node has effectively joined
// a cluster of existing members.
func (n *Node) IsMember() bool {
	return atomic.LoadUint32(&n.isMember) == 1
}

// canSubmitProposal defines if any more proposals
// could be submitted and processed.
func (n *Node) canSubmitProposal() bool {
	select {
	case <-n.stopped:
		return false
	default:
		return true
	}
}

// Saves a log entry to our Store
func (n *Node) saveToStorage(
	ctx context.Context,
	raftConfig *api.RaftConfig,
	hardState raftpb.HardState,
	entries []raftpb.Entry,
	snapshot raftpb.Snapshot,
) (err error) {

	if !raft.IsEmptySnap(snapshot) {
		if err := n.raftLogger.SaveSnapshot(snapshot); err != nil {
			return ErrApplySnapshot
		}
		if err := n.raftLogger.GC(snapshot.Metadata.Index, snapshot.Metadata.Term, raftConfig.KeepOldSnapshots); err != nil {
			log.G(ctx).WithError(err).Error("unable to clean old snapshots and WALs")
		}
		if err = n.raftStore.ApplySnapshot(snapshot); err != nil {
			return ErrApplySnapshot
		}
	}

	if err := n.raftLogger.SaveEntries(hardState, entries); err != nil {
		// TODO(aaronl): These error types should really wrap more
		// detailed errors.
		return ErrApplySnapshot
	}

	if len(entries) > 0 {
		lastIndex := entries[len(entries)-1].Index
		if lastIndex > n.writtenIndex {
			n.writtenIndex = lastIndex
		}
	}

	if err = n.raftStore.Append(entries); err != nil {
		return ErrAppendEntry
	}

	return nil
}

// Sends a series of messages to members in the raft
func (n *Node) send(ctx context.Context, messages []raftpb.Message) error {
	members := n.cluster.Members()

	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	for _, m := range messages {
		// Process locally
		if m.To == n.Config.ID {
			if err := n.raftNode.Step(ctx, m); err != nil {
				return err
			}
			continue
		}

		if m.Type == raftpb.MsgProp {
			// We don't forward proposals to the leader. Our
			// current architecture depends on only the leader
			// making proposals, so in-flight proposals can be
			// guaranteed not to conflict.
			continue
		}

		ch := make(chan struct{})

		n.asyncTasks.Add(1)
		go n.sendToMember(ctx, members, m, n.lastSendToMember[m.To], ch)

		n.lastSendToMember[m.To] = ch
	}

	return nil
}

func (n *Node) sendToMember(ctx context.Context, members map[uint64]*membership.Member, m raftpb.Message, lastSend <-chan struct{}, thisSend chan<- struct{}) {
	defer n.asyncTasks.Done()
	defer close(thisSend)

	ctx, cancel := context.WithTimeout(ctx, n.opts.SendTimeout)
	defer cancel()

	if lastSend != nil {
		select {
		case <-lastSend:
		case <-ctx.Done():
			return
		}
	}

	if n.cluster.IsIDRemoved(m.To) {
		// Should not send to removed members
		return
	}

	var (
		conn *membership.Member
	)
	if toMember, ok := members[m.To]; ok {
		conn = toMember
	} else {
		// If we are being asked to send to a member that's not in
		// our member list, that could indicate that the current leader
		// was added while we were offline. Try to resolve its address.
		log.G(ctx).Warningf("sending message to an unrecognized member ID %x", m.To)

		// Choose a random member
		var (
			queryMember *membership.Member
			id          uint64
		)
		for id, queryMember = range members {
			if id != n.Config.ID {
				break
			}
		}

		if queryMember == nil || queryMember.RaftID == n.Config.ID {
			log.G(ctx).Error("could not find cluster member to query for leader address")
			return
		}

		resp, err := api.NewRaftClient(queryMember.Conn).ResolveAddress(ctx, &api.ResolveAddressRequest{RaftID: m.To})
		if err != nil {
			log.G(ctx).WithError(err).Errorf("could not resolve address of member ID %x", m.To)
			return
		}
		conn, err = n.ConnectToMember(resp.Addr, n.opts.SendTimeout)
		if err != nil {
			log.G(ctx).WithError(err).Errorf("could connect to member ID %x at %s", m.To, resp.Addr)
			return
		}
		// The temporary connection is only used for this message.
		// Eventually, we should catch up and add a long-lived
		// connection to the member list.
		defer conn.Conn.Close()
	}

	_, err := api.NewRaftClient(conn.Conn).ProcessRaftMessage(ctx, &api.ProcessRaftMessageRequest{Message: &m})
	if err != nil {
		if grpc.Code(err) == codes.NotFound && grpc.ErrorDesc(err) == ErrMemberRemoved.Error() {
			n.removeRaftFunc()
		}
		if m.Type == raftpb.MsgSnap {
			n.raftNode.ReportSnapshot(m.To, raft.SnapshotFailure)
		}
		if !n.IsMember() {
			// node is removed from cluster or stopped
			return
		}
		n.raftNode.ReportUnreachable(m.To)

		lastSeenHost := n.cluster.LastSeenHost(m.To)
		if lastSeenHost != "" {
			// Check if address has changed
			officialHost, officialPort, _ := net.SplitHostPort(conn.Addr)
			if officialHost != lastSeenHost {
				reconnectAddr := net.JoinHostPort(lastSeenHost, officialPort)
				log.G(ctx).Warningf("detected address change for %x (%s -> %s)", m.To, conn.Addr, reconnectAddr)
				if err := n.handleAddressChange(ctx, conn, reconnectAddr); err != nil {
					log.G(ctx).WithError(err).Error("failed to hande address change")
				}
				return
			}
		}

		// Bounce the connection
		newConn, err := n.ConnectToMember(conn.Addr, 0)
		if err != nil {
			log.G(ctx).WithError(err).Errorf("could connect to member ID %x at %s", m.To, conn.Addr)
			return
		}
		err = n.cluster.ReplaceMemberConnection(m.To, conn, newConn, conn.Addr, false)
		if err != nil {
			log.G(ctx).WithError(err).Error("failed to replace connection to raft member")
			newConn.Conn.Close()
		}
	} else if m.Type == raftpb.MsgSnap {
		n.raftNode.ReportSnapshot(m.To, raft.SnapshotFinish)
	}
}

func (n *Node) handleAddressChange(ctx context.Context, member *membership.Member, reconnectAddr string) error {
	newConn, err := n.ConnectToMember(reconnectAddr, 0)
	if err != nil {
		return errors.Wrapf(err, "could connect to member ID %x at observed address %s", member.RaftID, reconnectAddr)
	}

	healthCtx, cancelHealth := context.WithTimeout(ctx, time.Duration(n.Config.ElectionTick)*n.opts.TickInterval)
	defer cancelHealth()

	if err := newConn.HealthCheck(healthCtx); err != nil {
		return errors.Wrapf(err, "%x failed health check at observed address %s", member.RaftID, reconnectAddr)
	}

	if err := n.cluster.ReplaceMemberConnection(member.RaftID, member, newConn, reconnectAddr, false); err != nil {
		newConn.Conn.Close()
		return errors.Wrap(err, "failed to replace connection to raft member")
	}

	// If we're the leader, write the address change to raft
	updateCtx, cancelUpdate := context.WithTimeout(ctx, time.Duration(n.Config.ElectionTick)*n.opts.TickInterval)
	defer cancelUpdate()
	if err := n.updateMember(updateCtx, reconnectAddr, member.RaftID, member.NodeID); err != nil {
		return errors.Wrap(err, "failed to update member address in raft")
	}

	return nil
}

type applyResult struct {
	resp proto.Message
	err  error
}

// processInternalRaftRequest sends a message to nodes participating
// in the raft to apply a log entry and then waits for it to be applied
// on the server. It will block until the update is performed, there is
// an error or until the raft node finalizes all the proposals on node
// shutdown.
func (n *Node) processInternalRaftRequest(ctx context.Context, r *api.InternalRaftRequest, cb func()) (proto.Message, error) {
	n.stopMu.RLock()
	if !n.canSubmitProposal() {
		n.stopMu.RUnlock()
		return nil, ErrStopped
	}
	n.waitProp.Add(1)
	defer n.waitProp.Done()
	n.stopMu.RUnlock()

	r.ID = n.reqIDGen.Next()

	// This must be derived from the context which is cancelled by stop()
	// to avoid a deadlock on shutdown.
	waitCtx, cancel := context.WithCancel(ctx)

	ch := n.wait.register(r.ID, cb, cancel)

	// Do this check after calling register to avoid a race.
	if atomic.LoadUint32(&n.signalledLeadership) != 1 {
		n.wait.cancel(r.ID)
		return nil, ErrLostLeadership
	}

	data, err := r.Marshal()
	if err != nil {
		n.wait.cancel(r.ID)
		return nil, err
	}

	if len(data) > store.MaxTransactionBytes {
		n.wait.cancel(r.ID)
		return nil, ErrRequestTooLarge
	}

	err = n.raftNode.Propose(waitCtx, data)
	if err != nil {
		n.wait.cancel(r.ID)
		return nil, err
	}

	select {
	case x := <-ch:
		res := x.(*applyResult)
		return res.resp, res.err
	case <-waitCtx.Done():
		n.wait.cancel(r.ID)
		return nil, ErrLostLeadership
	case <-ctx.Done():
		n.wait.cancel(r.ID)
		return nil, ctx.Err()
	}
}

// configure sends a configuration change through consensus and
// then waits for it to be applied to the server. It will block
// until the change is performed or there is an error.
func (n *Node) configure(ctx context.Context, cc raftpb.ConfChange) error {
	cc.ID = n.reqIDGen.Next()

	ctx, cancel := context.WithCancel(ctx)
	ch := n.wait.register(cc.ID, nil, cancel)

	if err := n.raftNode.ProposeConfChange(ctx, cc); err != nil {
		n.wait.cancel(cc.ID)
		return err
	}

	select {
	case x := <-ch:
		if err, ok := x.(error); ok {
			return err
		}
		if x != nil {
			log.G(ctx).Panic("raft: configuration change error, return type should always be error")
		}
		return nil
	case <-ctx.Done():
		n.wait.cancel(cc.ID)
		return ctx.Err()
	}
}

func (n *Node) processCommitted(ctx context.Context, entry raftpb.Entry) error {
	// Process a normal entry
	if entry.Type == raftpb.EntryNormal && entry.Data != nil {
		if err := n.processEntry(ctx, entry); err != nil {
			return err
		}
	}

	// Process a configuration change (add/remove node)
	if entry.Type == raftpb.EntryConfChange {
		n.processConfChange(ctx, entry)
	}

	n.appliedIndex = entry.Index
	return nil
}

func (n *Node) processEntry(ctx context.Context, entry raftpb.Entry) error {
	r := &api.InternalRaftRequest{}
	err := proto.Unmarshal(entry.Data, r)
	if err != nil {
		return err
	}

	if r.Action == nil {
		return nil
	}

	if !n.wait.trigger(r.ID, &applyResult{resp: r, err: nil}) {
		// There was no wait on this ID, meaning we don't have a
		// transaction in progress that would be committed to the
		// memory store by the "trigger" call. Either a different node
		// wrote this to raft, or we wrote it before losing the leader
		// position and cancelling the transaction. Create a new
		// transaction to commit the data.

		// It should not be possible for processInternalRaftRequest
		// to be running in this situation, but out of caution we
		// cancel any current invocations to avoid a deadlock.
		n.wait.cancelAll()

		err := n.memoryStore.ApplyStoreActions(r.Action)
		if err != nil {
			log.G(ctx).WithError(err).Error("failed to apply actions from raft")
		}
	}
	return nil
}

func (n *Node) processConfChange(ctx context.Context, entry raftpb.Entry) {
	var (
		err error
		cc  raftpb.ConfChange
	)

	if err := proto.Unmarshal(entry.Data, &cc); err != nil {
		n.wait.trigger(cc.ID, err)
	}

	if err := n.cluster.ValidateConfigurationChange(cc); err != nil {
		n.wait.trigger(cc.ID, err)
	}

	switch cc.Type {
	case raftpb.ConfChangeAddNode:
		err = n.applyAddNode(cc)
	case raftpb.ConfChangeUpdateNode:
		err = n.applyUpdateNode(ctx, cc)
	case raftpb.ConfChangeRemoveNode:
		err = n.applyRemoveNode(ctx, cc)
	}

	if err != nil {
		n.wait.trigger(cc.ID, err)
	}

	n.confState = *n.raftNode.ApplyConfChange(cc)
	n.wait.trigger(cc.ID, nil)
}

// applyAddNode is called when we receive a ConfChange
// from a member in the raft cluster, this adds a new
// node to the existing raft cluster
func (n *Node) applyAddNode(cc raftpb.ConfChange) error {
	member := &api.RaftMember{}
	err := proto.Unmarshal(cc.Context, member)
	if err != nil {
		return err
	}

	// ID must be non zero
	if member.RaftID == 0 {
		return nil
	}

	if err = n.registerNode(member); err != nil {
		return err
	}
	return nil
}

// applyUpdateNode is called when we receive a ConfChange from a member in the
// raft cluster which update the address of an existing node.
func (n *Node) applyUpdateNode(ctx context.Context, cc raftpb.ConfChange) error {
	newMember := &api.RaftMember{}
	err := proto.Unmarshal(cc.Context, newMember)
	if err != nil {
		return err
	}

	oldMember := n.cluster.GetMember(newMember.RaftID)

	if oldMember == nil {
		return ErrMemberUnknown
	}
	if oldMember.NodeID != newMember.NodeID {
		// Should never happen; this is a sanity check
		log.G(ctx).Errorf("node ID mismatch on node update (old: %x, new: %x)", oldMember.NodeID, newMember.NodeID)
		return errors.New("node ID mismatch match on node update")
	}

	if oldMember.Addr == newMember.Addr || oldMember.Conn == nil {
		// nothing to do
		return nil
	}

	newConn, err := n.ConnectToMember(newMember.Addr, 0)
	if err != nil {
		return errors.Errorf("could connect to member ID %x at %s: %v", newMember.RaftID, newMember.Addr, err)
	}
	if err := n.cluster.ReplaceMemberConnection(newMember.RaftID, oldMember, newConn, newMember.Addr, true); err != nil {
		newConn.Conn.Close()
		return err
	}

	return nil
}

// applyRemoveNode is called when we receive a ConfChange
// from a member in the raft cluster, this removes a node
// from the existing raft cluster
func (n *Node) applyRemoveNode(ctx context.Context, cc raftpb.ConfChange) (err error) {
	// If the node from where the remove is issued is
	// a follower and the leader steps down, Campaign
	// to be the leader.

	if cc.NodeID == n.leader() && !n.isLeader() {
		if err = n.raftNode.Campaign(ctx); err != nil {
			return err
		}
	}

	if cc.NodeID == n.Config.ID {
		n.removeRaftFunc()

		// wait the commit ack to be sent before closing connection
		n.asyncTasks.Wait()

		// if there are only 2 nodes in the cluster, and leader is leaving
		// before closing the connection, leader has to ensure that follower gets
		// noticed about this raft conf change commit. Otherwise, follower would
		// assume there are still 2 nodes in the cluster and won't get elected
		// into the leader by acquiring the majority (2 nodes)

		// while n.asyncTasks.Wait() could be helpful in this case
		// it's the best-effort strategy, because this send could be fail due to some errors (such as time limit exceeds)
		// TODO(Runshen Zhu): use leadership transfer to solve this case, after vendoring raft 3.0+
	}

	return n.cluster.RemoveMember(cc.NodeID)
}

// ConnectToMember returns a member object with an initialized
// connection to communicate with other raft members
func (n *Node) ConnectToMember(addr string, timeout time.Duration) (*membership.Member, error) {
	conn, err := dial(addr, "tcp", n.opts.TLSCredentials, timeout)
	if err != nil {
		return nil, err
	}

	return &membership.Member{
		Conn: conn,
	}, nil
}

// SubscribeLeadership returns channel to which events about leadership change
// will be sent in form of raft.LeadershipState. Also cancel func is returned -
// it should be called when listener is no longer interested in events.
func (n *Node) SubscribeLeadership() (q chan events.Event, cancel func()) {
	return n.leadershipBroadcast.Watch()
}

// createConfigChangeEnts creates a series of Raft entries (i.e.
// EntryConfChange) to remove the set of given IDs from the cluster. The ID
// `self` is _not_ removed, even if present in the set.
// If `self` is not inside the given ids, it creates a Raft entry to add a
// default member with the given `self`.
func createConfigChangeEnts(ids []uint64, self uint64, term, index uint64) []raftpb.Entry {
	var ents []raftpb.Entry
	next := index + 1
	found := false
	for _, id := range ids {
		if id == self {
			found = true
			continue
		}
		cc := &raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: id,
		}
		data, err := cc.Marshal()
		if err != nil {
			log.L.WithError(err).Panic("marshal configuration change should never fail")
		}
		e := raftpb.Entry{
			Type:  raftpb.EntryConfChange,
			Data:  data,
			Term:  term,
			Index: next,
		}
		ents = append(ents, e)
		next++
	}
	if !found {
		node := &api.RaftMember{RaftID: self}
		meta, err := node.Marshal()
		if err != nil {
			log.L.WithError(err).Panic("marshal member should never fail")
		}
		cc := &raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  self,
			Context: meta,
		}
		data, err := cc.Marshal()
		if err != nil {
			log.L.WithError(err).Panic("marshal configuration change should never fail")
		}
		e := raftpb.Entry{
			Type:  raftpb.EntryConfChange,
			Data:  data,
			Term:  term,
			Index: next,
		}
		ents = append(ents, e)
	}
	return ents
}

// getIDs returns an ordered set of IDs included in the given snapshot and
// the entries. The given snapshot/entries can contain two kinds of
// ID-related entry:
// - ConfChangeAddNode, in which case the contained ID will be added into the set.
// - ConfChangeRemoveNode, in which case the contained ID will be removed from the set.
func getIDs(snap *raftpb.Snapshot, ents []raftpb.Entry) []uint64 {
	ids := make(map[uint64]bool)
	if snap != nil {
		for _, id := range snap.Metadata.ConfState.Nodes {
			ids[id] = true
		}
	}
	for _, e := range ents {
		if e.Type != raftpb.EntryConfChange {
			continue
		}
		if snap != nil && e.Index < snap.Metadata.Index {
			continue
		}
		var cc raftpb.ConfChange
		if err := cc.Unmarshal(e.Data); err != nil {
			log.L.WithError(err).Panic("unmarshal configuration change should never fail")
		}
		switch cc.Type {
		case raftpb.ConfChangeAddNode:
			ids[cc.NodeID] = true
		case raftpb.ConfChangeRemoveNode:
			delete(ids, cc.NodeID)
		case raftpb.ConfChangeUpdateNode:
			// do nothing
		default:
			log.L.Panic("ConfChange Type should be either ConfChangeAddNode, or ConfChangeRemoveNode, or ConfChangeUpdateNode!")
		}
	}
	var sids []uint64
	for id := range ids {
		sids = append(sids, id)
	}
	return sids
}
