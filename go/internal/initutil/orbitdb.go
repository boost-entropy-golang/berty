package initutil

import (
	datastore "github.com/ipfs/go-datastore"

	"berty.tech/berty/v2/go/internal/datastoreutil"
	"berty.tech/berty/v2/go/pkg/errcode"
	"berty.tech/go-orbit-db/baseorbitdb"
	"berty.tech/go-orbit-db/pubsub/directchannel"
	"berty.tech/go-orbit-db/pubsub/pubsubraw"
	"berty.tech/weshnet"
	"berty.tech/weshnet/pkg/cryptoutil"
	"berty.tech/weshnet/pkg/ipfsutil"
	"berty.tech/weshnet/pkg/rendezvous"
)

const (
	DefaultBertyGroupMetadataStoreType = "berty_group_metadata"
	DefaultBertyGroupMessageStoreType  = "berty_group_messages"
)

func (m *Manager) GetRotationInterval() (rp *rendezvous.RotationInterval, err error) {
	m.mutex.Lock()
	rp, err = m.getRotationInterval()
	m.mutex.Unlock()
	return
}

func (m *Manager) getRotationInterval() (*rendezvous.RotationInterval, error) {
	if m.Node.Protocol.rotationInterval == nil {
		rendezvousRotationBase, err := m.GetRendezvousRotationBase()
		if err != nil {
			return nil, errcode.ErrDeserialization.Wrap(err)
		}
		m.Node.Protocol.rotationInterval = rendezvous.NewRotationInterval(rendezvousRotationBase)
	}

	return m.Node.Protocol.rotationInterval, nil
}

func (m *Manager) getOrbitDB() (*weshnet.WeshOrbitDB, error) {
	m.applyDefaults()

	if m.Node.Protocol.orbitDB != nil {
		return m.Node.Protocol.orbitDB, nil
	}

	ipfs, node, err := m.getLocalIPFS()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	logger, err := m.getLogger()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	rootDS, err := m.getRootDatastore()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	var (
		deviceDS = ipfsutil.NewDatastoreKeystore(datastoreutil.NewNamespacedDatastore(rootDS, datastore.NewKey(weshnet.NamespaceDeviceKeystore)))
		deviceKS = cryptoutil.NewDeviceKeystore(deviceDS, nil)
		cache    = weshnet.NewOrbitDatastoreCache(rootDS)
	)

	rp, err := m.getRotationInterval()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	opts := &weshnet.NewOrbitDBOptions{
		GroupMetadataStoreType: DefaultBertyGroupMetadataStoreType,
		GroupMessageStoreType:  DefaultBertyGroupMessageStoreType,
		NewOrbitDBOptions: baseorbitdb.NewOrbitDBOptions{
			Cache:                cache,
			Logger:               logger,
			DirectChannelFactory: directchannel.InitDirectChannelFactory(logger.Named("odb-dc"), node.PeerHost),
		},
		Datastore:        rootDS,
		DeviceKeystore:   deviceKS,
		RotationInterval: rp,
	}

	if node.PubSub != nil {
		self, err := ipfs.Key().Self(m.getContext())
		if err != nil {
			return nil, errcode.TODO.Wrap(err)
		}

		opts.PubSub = pubsubraw.NewPubSub(node.PubSub, self.ID(), opts.Logger, nil)
	}

	odb, err := weshnet.NewWeshOrbitDB(m.getContext(), ipfs, opts)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	m.Node.Protocol.orbitDB = odb

	return odb, nil
}

func (m *Manager) GetOrbitDB() (*weshnet.WeshOrbitDB, error) {
	defer m.prepareForGetter()()

	return m.getOrbitDB()
}
