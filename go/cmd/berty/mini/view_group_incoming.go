package mini

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"berty.tech/berty/v2/go/pkg/errcode"
	"berty.tech/weshnet/pkg/protocoltypes"
)

func handlerAccountGroupJoined(ctx context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountGroupJoined{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("joined a group"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	v.v.AddContextGroup(ctx, casted.Group)
	v.v.recomputeChannelList(false)

	return nil
}

func handlerGroupDeviceChainKeyAdded(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.GroupDeviceChainKeyAdded{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	v.muAggregates.Lock()
	v.secrets[string(casted.DevicePK)] = casted
	v.muAggregates.Unlock()

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("has exchanged a secret"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	return nil
}

func handlerGroupMemberDeviceAdded(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.GroupMemberDeviceAdded{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	v.muAggregates.Lock()
	v.devices[string(casted.DevicePK)] = casted
	v.muAggregates.Unlock()

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("new device joined the group"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	return nil
}

func handlerAccountContactRequestOutgoingSent(ctx context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountContactRequestOutgoingSent{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("outgoing contact request sent"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	gInfo, err := v.v.protocol.GroupInfo(ctx, &protocoltypes.GroupInfo_Request{
		ContactPK: casted.ContactPK,
	})
	if err != nil {
		return err
	}

	v.v.lock.Lock()
	if _, hasValue := v.v.contactStates[string(casted.ContactPK)]; !hasValue || !isHistory {
		v.v.contactStates[string(casted.ContactPK)] = protocoltypes.ContactStateAdded
	}

	v.v.lock.Unlock()

	v.v.AddContextGroup(ctx, gInfo.Group)
	v.v.recomputeChannelList(true)

	return nil
}

func handlerAccountGroupLeft(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountGroupLeft{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("left group %s", base64.StdEncoding.EncodeToString(casted.GroupPK))),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	return nil
}

func handlerAccountContactRequestIncomingReceived(ctx context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountContactRequestIncomingReceived{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	name := string(casted.ContactMetadata)

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("incoming request received %s, type /contact accept %s (alt. /contact discard <id>)", name, base64.StdEncoding.EncodeToString(casted.ContactPK))),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	v.v.lock.Lock()
	if _, hasValue := v.v.contactStates[string(casted.ContactPK)]; !hasValue || !isHistory {
		v.v.contactStates[string(casted.ContactPK)] = protocoltypes.ContactStateReceived
	}
	v.v.lock.Unlock()

	gInfo, err := v.v.protocol.GroupInfo(ctx, &protocoltypes.GroupInfo_Request{
		ContactPK: casted.ContactPK,
	})

	if err == nil {
		v.v.lock.Lock()
		if _, hasValue := v.v.contactNames[string(gInfo.Group.PublicKey)]; (!hasValue || !isHistory) && len(casted.ContactMetadata) > 0 {
			v.v.contactNames[string(gInfo.Group.PublicKey)] = string(casted.ContactMetadata)
		}
		v.v.lock.Unlock()
	}

	return nil
}

func handlerAccountContactRequestIncomingDiscarded(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountContactRequestIncomingDiscarded{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("incoming request discarded, contact: %s", base64.StdEncoding.EncodeToString(casted.ContactPK))),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	v.v.lock.Lock()
	if _, hasValue := v.v.contactStates[string(casted.ContactPK)]; !hasValue || !isHistory {
		v.v.contactStates[string(casted.ContactPK)] = protocoltypes.ContactStateRemoved
	}
	v.v.lock.Unlock()

	return nil
}

func handlerMultiMemberGroupInitialMemberAnnounced(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.MultiMemberGroupInitialMemberAnnounced{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("member claimed group ownership"),
		sender:      casted.MemberPK,
	}, e, v, isHistory)

	return nil
}

func handlerAccountContactRequestOutgoingEnqueued(ctx context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountContactRequestOutgoingEnqueued{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}
	if casted.Contact == nil {
		return errcode.ErrInvalidInput
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("outgoing contact request enqueued (%s)", base64.StdEncoding.EncodeToString(casted.Contact.PK))),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("fake request on the other end by typing `/contact received` with the value of `/contact share`"),
		sender:      casted.DevicePK,
	}, nil, v, false)

	v.v.lock.Lock()
	if _, hasValue := v.v.contactStates[string(casted.Contact.PK)]; !hasValue || !isHistory {
		v.v.contactStates[string(casted.Contact.PK)] = protocoltypes.ContactStateToRequest
	}
	v.v.lock.Unlock()

	gInfo, err := v.v.protocol.GroupInfo(ctx, &protocoltypes.GroupInfo_Request{
		ContactPK: casted.Contact.PK,
	})

	if err == nil {
		v.v.AddContextGroup(ctx, gInfo.Group)
		v.v.recomputeChannelList(true)

		v.v.lock.Lock()
		if _, hasValue := v.v.contactNames[string(gInfo.Group.PublicKey)]; (!hasValue || !isHistory) && len(casted.Contact.Metadata) > 0 {
			v.v.contactNames[string(gInfo.Group.PublicKey)] = string(casted.Contact.Metadata)
		}

		v.v.lock.Unlock()
	}

	return nil
}

func handlerContactAliasKeyAdded(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.ContactAliasKeyAdded{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("contact alias public key received"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	return nil
}

func handlerMultiMemberGroupAliasResolverAdded(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.MultiMemberGroupAliasResolverAdded{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("contact alias proof received"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	return nil
}

func handlerAccountContactRequestIncomingAccepted(ctx context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountContactRequestOutgoingSent{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte("incoming contact request accepted"),
		sender:      casted.DevicePK,
	}, e, v, isHistory)

	gInfo, err := v.v.protocol.GroupInfo(ctx, &protocoltypes.GroupInfo_Request{
		ContactPK: casted.ContactPK,
	})
	if err != nil {
		return err
	}

	v.v.lock.Lock()
	if _, hasValue := v.v.contactStates[string(casted.ContactPK)]; !hasValue || !isHistory {
		v.v.contactStates[string(casted.ContactPK)] = protocoltypes.ContactStateAdded
	}
	v.v.lock.Unlock()

	v.v.AddContextGroup(ctx, gInfo.Group)
	v.v.recomputeChannelList(false)

	return nil
}

func handlerNoop(_ context.Context, _ *groupView, _ *protocoltypes.GroupMetadataEvent, _ bool) error {
	return nil
}

func groupDeviceStatusHandler(logger *zap.Logger, v *groupView, e *protocoltypes.GroupDeviceStatus_Reply) {
	var payload string

	switch t := e.GetType(); t {
	case protocoltypes.TypePeerConnected:
		event := &protocoltypes.GroupDeviceStatus_Reply_PeerConnected{}
		if err := event.Unmarshal(e.GetEvent()); err != nil {
			logger.Error("unmarshal error", zap.Error(err))
			return
		}

		activeAddr := "<unknown>"
		if maddrs := event.GetMaddrs(); len(maddrs) > 0 {
			activeAddr = maddrs[0]
		}

		activeTransport := "<unknown>"
		if tpts := event.GetTransports(); len(tpts) > 0 {
			activeTransport = tpts[0].String()
		}

		payload = fmt.Sprintf("device status updated: connected <%.15s> on: %s(%s)", event.GetPeerID(), activeAddr, activeTransport)

	case protocoltypes.TypePeerDisconnected:
		event := &protocoltypes.GroupDeviceStatus_Reply_PeerDisconnected{}
		if err := event.Unmarshal(e.GetEvent()); err != nil {
			logger.Error("unmarshal error", zap.Error(err))
			return
		}
		payload = fmt.Sprintf("device status updated: left <%.15s>", event.GetPeerID())

	case protocoltypes.TypePeerReconnecting:
		event := &protocoltypes.GroupDeviceStatus_Reply_PeerReconnecting{}
		if err := event.Unmarshal(e.GetEvent()); err != nil {
			logger.Error("unmarshal error", zap.Error(err))
			return
		}
		payload = fmt.Sprintf("device status updated: reconnecting <%.15s>", event.GetPeerID())
	default:
		logger.Warn("unknow group device status event received")
		return
	}

	v.messages.Append(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(payload),
	})
}

func metadataEventHandler(ctx context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool, logger *zap.Logger) {
	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("event type: %s", e.Metadata.EventType.String())),
	}, e, v, isHistory)

	actions := map[protocoltypes.EventType]func(context.Context, *groupView, *protocoltypes.GroupMetadataEvent, bool) error{
		protocoltypes.EventTypeAccountContactBlocked:                  nil, // do it later
		protocoltypes.EventTypeAccountContactRequestDisabled:          handlerNoop,
		protocoltypes.EventTypeAccountContactRequestEnabled:           handlerNoop,
		protocoltypes.EventTypeAccountContactRequestIncomingAccepted:  handlerAccountContactRequestIncomingAccepted,
		protocoltypes.EventTypeAccountContactRequestIncomingDiscarded: handlerAccountContactRequestIncomingDiscarded,
		protocoltypes.EventTypeAccountContactRequestIncomingReceived:  handlerAccountContactRequestIncomingReceived,
		protocoltypes.EventTypeAccountContactRequestOutgoingEnqueued:  handlerAccountContactRequestOutgoingEnqueued,
		protocoltypes.EventTypeAccountContactRequestOutgoingSent:      handlerAccountContactRequestOutgoingSent,
		protocoltypes.EventTypeAccountContactRequestReferenceReset:    handlerNoop,
		protocoltypes.EventTypeAccountContactUnblocked:                nil, // do it later
		protocoltypes.EventTypeAccountGroupJoined:                     handlerAccountGroupJoined,
		protocoltypes.EventTypeAccountGroupLeft:                       handlerAccountGroupLeft,
		protocoltypes.EventTypeContactAliasKeyAdded:                   handlerContactAliasKeyAdded,
		protocoltypes.EventTypeGroupDeviceChainKeyAdded:               handlerGroupDeviceChainKeyAdded,
		protocoltypes.EventTypeGroupMemberDeviceAdded:                 handlerGroupMemberDeviceAdded,
		protocoltypes.EventTypeGroupMetadataPayloadSent:               nil, // do it later
		protocoltypes.EventTypeMultiMemberGroupAdminRoleGranted:       nil, // do it later
		protocoltypes.EventTypeMultiMemberGroupAliasResolverAdded:     handlerMultiMemberGroupAliasResolverAdded,
		protocoltypes.EventTypeMultiMemberGroupInitialMemberAnnounced: handlerMultiMemberGroupInitialMemberAnnounced,
		protocoltypes.EventTypeAccountServiceTokenAdded:               handlerAccountServiceTokenAdded,
	}
	logger.Debug("metadataEventHandler", zap.Stringer("event-type", e.Metadata.EventType))

	action, ok := actions[e.Metadata.EventType]
	if !ok || action == nil {
		v.messages.AppendErr(fmt.Errorf("action handler for %s not found", e.Metadata.EventType.String()))
		v.addBadge()
		return
	}

	if err := action(ctx, v, e, isHistory); err != nil {
		v.messages.AppendErr(fmt.Errorf("error while handling metadata event (type: %s): %w", e.Metadata.EventType.String(), err))
		v.addBadge()
	}
}

func handlerAccountServiceTokenAdded(_ context.Context, v *groupView, e *protocoltypes.GroupMetadataEvent, isHistory bool) error {
	casted := &protocoltypes.AccountServiceTokenAdded{}
	if err := casted.Unmarshal(e.Event); err != nil {
		return err
	}

	addToBuffer(&historyMessage{
		messageType: messageTypeMeta,
		payload:     []byte(fmt.Sprintf("service token registered for account (%s: auth via %s)", casted.ServiceToken.TokenID(), casted.ServiceToken.AuthenticationURL)),
	}, e, v, isHistory)

	for _, s := range casted.ServiceToken.SupportedServices {
		addToBuffer(&historyMessage{
			messageType: messageTypeMeta,
			payload:     []byte(fmt.Sprintf(" - %s, %s", s.ServiceType, s.ServiceEndpoint)),
		}, e, v, isHistory)
	}

	return nil
}

func addToBuffer(evt *historyMessage, _ *protocoltypes.GroupMetadataEvent, v *groupView, isHistory bool) {
	if isHistory {
		v.messages.Prepend(evt, time.Time{})
	} else {
		v.messages.Append(evt)
		v.addBadge()
	}
}

func (v *groupView) addBadge() {
	// Display unread badge
	recompute := false
	v.v.lock.Lock()
	if v.v.selectedGroupView != v {
		atomic.StoreInt32(&v.hasNew, 1)
		recompute = true
	}
	v.v.lock.Unlock()

	if recompute {
		v.v.recomputeChannelList(true)
	}
}
