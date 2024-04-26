// Code generated by counterfeiter. DO NOT EDIT.
package typesfakes

import (
	"sync"

	"github.com/livekit/livekit-server/pkg/rtc/types"
	"github.com/livekit/protocol/livekit"
)

type FakeRoom struct {
	GetLocalParticipantsStub        func() []types.LocalParticipant
	getLocalParticipantsMutex       sync.RWMutex
	getLocalParticipantsArgsForCall []struct {
	}
	getLocalParticipantsReturns struct {
		result1 []types.LocalParticipant
	}
	getLocalParticipantsReturnsOnCall map[int]struct {
		result1 []types.LocalParticipant
	}
	IDStub        func() livekit.RoomID
	iDMutex       sync.RWMutex
	iDArgsForCall []struct {
	}
	iDReturns struct {
		result1 livekit.RoomID
	}
	iDReturnsOnCall map[int]struct {
		result1 livekit.RoomID
	}
	NameStub        func() livekit.RoomName
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 livekit.RoomName
	}
	nameReturnsOnCall map[int]struct {
		result1 livekit.RoomName
	}
	RemoveParticipantStub        func(livekit.ParticipantIdentity, livekit.ParticipantID, types.ParticipantCloseReason)
	removeParticipantMutex       sync.RWMutex
	removeParticipantArgsForCall []struct {
		arg1 livekit.ParticipantIdentity
		arg2 livekit.ParticipantID
		arg3 types.ParticipantCloseReason
	}
	ResolveMediaTrackForSubscriberStub        func(livekit.ParticipantIdentity, livekit.TrackID) types.MediaResolverResult
	resolveMediaTrackForSubscriberMutex       sync.RWMutex
	resolveMediaTrackForSubscriberArgsForCall []struct {
		arg1 livekit.ParticipantIdentity
		arg2 livekit.TrackID
	}
	resolveMediaTrackForSubscriberReturns struct {
		result1 types.MediaResolverResult
	}
	resolveMediaTrackForSubscriberReturnsOnCall map[int]struct {
		result1 types.MediaResolverResult
	}
	SimulateScenarioStub        func(types.LocalParticipant, *livekit.SimulateScenario) error
	simulateScenarioMutex       sync.RWMutex
	simulateScenarioArgsForCall []struct {
		arg1 types.LocalParticipant
		arg2 *livekit.SimulateScenario
	}
	simulateScenarioReturns struct {
		result1 error
	}
	simulateScenarioReturnsOnCall map[int]struct {
		result1 error
	}
	SyncStateStub        func(types.LocalParticipant, *livekit.SyncState) error
	syncStateMutex       sync.RWMutex
	syncStateArgsForCall []struct {
		arg1 types.LocalParticipant
		arg2 *livekit.SyncState
	}
	syncStateReturns struct {
		result1 error
	}
	syncStateReturnsOnCall map[int]struct {
		result1 error
	}
	UpdateParticipantMetadataStub        func(types.LocalParticipant, string, string)
	updateParticipantMetadataMutex       sync.RWMutex
	updateParticipantMetadataArgsForCall []struct {
		arg1 types.LocalParticipant
		arg2 string
		arg3 string
	}
	UpdateSubscriptionPermissionStub        func(types.LocalParticipant, *livekit.SubscriptionPermission) error
	updateSubscriptionPermissionMutex       sync.RWMutex
	updateSubscriptionPermissionArgsForCall []struct {
		arg1 types.LocalParticipant
		arg2 *livekit.SubscriptionPermission
	}
	updateSubscriptionPermissionReturns struct {
		result1 error
	}
	updateSubscriptionPermissionReturnsOnCall map[int]struct {
		result1 error
	}
	UpdateSubscriptionsStub        func(types.LocalParticipant, []livekit.TrackID, []*livekit.ParticipantTracks, bool)
	updateSubscriptionsMutex       sync.RWMutex
	updateSubscriptionsArgsForCall []struct {
		arg1 types.LocalParticipant
		arg2 []livekit.TrackID
		arg3 []*livekit.ParticipantTracks
		arg4 bool
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRoom) GetLocalParticipants() []types.LocalParticipant {
	fake.getLocalParticipantsMutex.Lock()
	ret, specificReturn := fake.getLocalParticipantsReturnsOnCall[len(fake.getLocalParticipantsArgsForCall)]
	fake.getLocalParticipantsArgsForCall = append(fake.getLocalParticipantsArgsForCall, struct {
	}{})
	stub := fake.GetLocalParticipantsStub
	fakeReturns := fake.getLocalParticipantsReturns
	fake.recordInvocation("GetLocalParticipants", []interface{}{})
	fake.getLocalParticipantsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) GetLocalParticipantsCallCount() int {
	fake.getLocalParticipantsMutex.RLock()
	defer fake.getLocalParticipantsMutex.RUnlock()
	return len(fake.getLocalParticipantsArgsForCall)
}

func (fake *FakeRoom) GetLocalParticipantsCalls(stub func() []types.LocalParticipant) {
	fake.getLocalParticipantsMutex.Lock()
	defer fake.getLocalParticipantsMutex.Unlock()
	fake.GetLocalParticipantsStub = stub
}

func (fake *FakeRoom) GetLocalParticipantsReturns(result1 []types.LocalParticipant) {
	fake.getLocalParticipantsMutex.Lock()
	defer fake.getLocalParticipantsMutex.Unlock()
	fake.GetLocalParticipantsStub = nil
	fake.getLocalParticipantsReturns = struct {
		result1 []types.LocalParticipant
	}{result1}
}

func (fake *FakeRoom) GetLocalParticipantsReturnsOnCall(i int, result1 []types.LocalParticipant) {
	fake.getLocalParticipantsMutex.Lock()
	defer fake.getLocalParticipantsMutex.Unlock()
	fake.GetLocalParticipantsStub = nil
	if fake.getLocalParticipantsReturnsOnCall == nil {
		fake.getLocalParticipantsReturnsOnCall = make(map[int]struct {
			result1 []types.LocalParticipant
		})
	}
	fake.getLocalParticipantsReturnsOnCall[i] = struct {
		result1 []types.LocalParticipant
	}{result1}
}

func (fake *FakeRoom) ID() livekit.RoomID {
	fake.iDMutex.Lock()
	ret, specificReturn := fake.iDReturnsOnCall[len(fake.iDArgsForCall)]
	fake.iDArgsForCall = append(fake.iDArgsForCall, struct {
	}{})
	stub := fake.IDStub
	fakeReturns := fake.iDReturns
	fake.recordInvocation("ID", []interface{}{})
	fake.iDMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) IDCallCount() int {
	fake.iDMutex.RLock()
	defer fake.iDMutex.RUnlock()
	return len(fake.iDArgsForCall)
}

func (fake *FakeRoom) IDCalls(stub func() livekit.RoomID) {
	fake.iDMutex.Lock()
	defer fake.iDMutex.Unlock()
	fake.IDStub = stub
}

func (fake *FakeRoom) IDReturns(result1 livekit.RoomID) {
	fake.iDMutex.Lock()
	defer fake.iDMutex.Unlock()
	fake.IDStub = nil
	fake.iDReturns = struct {
		result1 livekit.RoomID
	}{result1}
}

func (fake *FakeRoom) IDReturnsOnCall(i int, result1 livekit.RoomID) {
	fake.iDMutex.Lock()
	defer fake.iDMutex.Unlock()
	fake.IDStub = nil
	if fake.iDReturnsOnCall == nil {
		fake.iDReturnsOnCall = make(map[int]struct {
			result1 livekit.RoomID
		})
	}
	fake.iDReturnsOnCall[i] = struct {
		result1 livekit.RoomID
	}{result1}
}

func (fake *FakeRoom) Name() livekit.RoomName {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct {
	}{})
	stub := fake.NameStub
	fakeReturns := fake.nameReturns
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeRoom) NameCalls(stub func() livekit.RoomName) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *FakeRoom) NameReturns(result1 livekit.RoomName) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 livekit.RoomName
	}{result1}
}

func (fake *FakeRoom) NameReturnsOnCall(i int, result1 livekit.RoomName) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 livekit.RoomName
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 livekit.RoomName
	}{result1}
}

func (fake *FakeRoom) RemoveParticipant(arg1 livekit.ParticipantIdentity, arg2 livekit.ParticipantID, arg3 types.ParticipantCloseReason) {
	fake.removeParticipantMutex.Lock()
	fake.removeParticipantArgsForCall = append(fake.removeParticipantArgsForCall, struct {
		arg1 livekit.ParticipantIdentity
		arg2 livekit.ParticipantID
		arg3 types.ParticipantCloseReason
	}{arg1, arg2, arg3})
	stub := fake.RemoveParticipantStub
	fake.recordInvocation("RemoveParticipant", []interface{}{arg1, arg2, arg3})
	fake.removeParticipantMutex.Unlock()
	if stub != nil {
		fake.RemoveParticipantStub(arg1, arg2, arg3)
	}
}

func (fake *FakeRoom) RemoveParticipantCallCount() int {
	fake.removeParticipantMutex.RLock()
	defer fake.removeParticipantMutex.RUnlock()
	return len(fake.removeParticipantArgsForCall)
}

func (fake *FakeRoom) RemoveParticipantCalls(stub func(livekit.ParticipantIdentity, livekit.ParticipantID, types.ParticipantCloseReason)) {
	fake.removeParticipantMutex.Lock()
	defer fake.removeParticipantMutex.Unlock()
	fake.RemoveParticipantStub = stub
}

func (fake *FakeRoom) RemoveParticipantArgsForCall(i int) (livekit.ParticipantIdentity, livekit.ParticipantID, types.ParticipantCloseReason) {
	fake.removeParticipantMutex.RLock()
	defer fake.removeParticipantMutex.RUnlock()
	argsForCall := fake.removeParticipantArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeRoom) ResolveMediaTrackForSubscriber(arg1 livekit.ParticipantIdentity, arg2 livekit.TrackID) types.MediaResolverResult {
	fake.resolveMediaTrackForSubscriberMutex.Lock()
	ret, specificReturn := fake.resolveMediaTrackForSubscriberReturnsOnCall[len(fake.resolveMediaTrackForSubscriberArgsForCall)]
	fake.resolveMediaTrackForSubscriberArgsForCall = append(fake.resolveMediaTrackForSubscriberArgsForCall, struct {
		arg1 livekit.ParticipantIdentity
		arg2 livekit.TrackID
	}{arg1, arg2})
	stub := fake.ResolveMediaTrackForSubscriberStub
	fakeReturns := fake.resolveMediaTrackForSubscriberReturns
	fake.recordInvocation("ResolveMediaTrackForSubscriber", []interface{}{arg1, arg2})
	fake.resolveMediaTrackForSubscriberMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) ResolveMediaTrackForSubscriberCallCount() int {
	fake.resolveMediaTrackForSubscriberMutex.RLock()
	defer fake.resolveMediaTrackForSubscriberMutex.RUnlock()
	return len(fake.resolveMediaTrackForSubscriberArgsForCall)
}

func (fake *FakeRoom) ResolveMediaTrackForSubscriberCalls(stub func(livekit.ParticipantIdentity, livekit.TrackID) types.MediaResolverResult) {
	fake.resolveMediaTrackForSubscriberMutex.Lock()
	defer fake.resolveMediaTrackForSubscriberMutex.Unlock()
	fake.ResolveMediaTrackForSubscriberStub = stub
}

func (fake *FakeRoom) ResolveMediaTrackForSubscriberArgsForCall(i int) (livekit.ParticipantIdentity, livekit.TrackID) {
	fake.resolveMediaTrackForSubscriberMutex.RLock()
	defer fake.resolveMediaTrackForSubscriberMutex.RUnlock()
	argsForCall := fake.resolveMediaTrackForSubscriberArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRoom) ResolveMediaTrackForSubscriberReturns(result1 types.MediaResolverResult) {
	fake.resolveMediaTrackForSubscriberMutex.Lock()
	defer fake.resolveMediaTrackForSubscriberMutex.Unlock()
	fake.ResolveMediaTrackForSubscriberStub = nil
	fake.resolveMediaTrackForSubscriberReturns = struct {
		result1 types.MediaResolverResult
	}{result1}
}

func (fake *FakeRoom) ResolveMediaTrackForSubscriberReturnsOnCall(i int, result1 types.MediaResolverResult) {
	fake.resolveMediaTrackForSubscriberMutex.Lock()
	defer fake.resolveMediaTrackForSubscriberMutex.Unlock()
	fake.ResolveMediaTrackForSubscriberStub = nil
	if fake.resolveMediaTrackForSubscriberReturnsOnCall == nil {
		fake.resolveMediaTrackForSubscriberReturnsOnCall = make(map[int]struct {
			result1 types.MediaResolverResult
		})
	}
	fake.resolveMediaTrackForSubscriberReturnsOnCall[i] = struct {
		result1 types.MediaResolverResult
	}{result1}
}

func (fake *FakeRoom) SimulateScenario(arg1 types.LocalParticipant, arg2 *livekit.SimulateScenario) error {
	fake.simulateScenarioMutex.Lock()
	ret, specificReturn := fake.simulateScenarioReturnsOnCall[len(fake.simulateScenarioArgsForCall)]
	fake.simulateScenarioArgsForCall = append(fake.simulateScenarioArgsForCall, struct {
		arg1 types.LocalParticipant
		arg2 *livekit.SimulateScenario
	}{arg1, arg2})
	stub := fake.SimulateScenarioStub
	fakeReturns := fake.simulateScenarioReturns
	fake.recordInvocation("SimulateScenario", []interface{}{arg1, arg2})
	fake.simulateScenarioMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) SimulateScenarioCallCount() int {
	fake.simulateScenarioMutex.RLock()
	defer fake.simulateScenarioMutex.RUnlock()
	return len(fake.simulateScenarioArgsForCall)
}

func (fake *FakeRoom) SimulateScenarioCalls(stub func(types.LocalParticipant, *livekit.SimulateScenario) error) {
	fake.simulateScenarioMutex.Lock()
	defer fake.simulateScenarioMutex.Unlock()
	fake.SimulateScenarioStub = stub
}

func (fake *FakeRoom) SimulateScenarioArgsForCall(i int) (types.LocalParticipant, *livekit.SimulateScenario) {
	fake.simulateScenarioMutex.RLock()
	defer fake.simulateScenarioMutex.RUnlock()
	argsForCall := fake.simulateScenarioArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRoom) SimulateScenarioReturns(result1 error) {
	fake.simulateScenarioMutex.Lock()
	defer fake.simulateScenarioMutex.Unlock()
	fake.SimulateScenarioStub = nil
	fake.simulateScenarioReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRoom) SimulateScenarioReturnsOnCall(i int, result1 error) {
	fake.simulateScenarioMutex.Lock()
	defer fake.simulateScenarioMutex.Unlock()
	fake.SimulateScenarioStub = nil
	if fake.simulateScenarioReturnsOnCall == nil {
		fake.simulateScenarioReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.simulateScenarioReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRoom) SyncState(arg1 types.LocalParticipant, arg2 *livekit.SyncState) error {
	fake.syncStateMutex.Lock()
	ret, specificReturn := fake.syncStateReturnsOnCall[len(fake.syncStateArgsForCall)]
	fake.syncStateArgsForCall = append(fake.syncStateArgsForCall, struct {
		arg1 types.LocalParticipant
		arg2 *livekit.SyncState
	}{arg1, arg2})
	stub := fake.SyncStateStub
	fakeReturns := fake.syncStateReturns
	fake.recordInvocation("SyncState", []interface{}{arg1, arg2})
	fake.syncStateMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) SyncStateCallCount() int {
	fake.syncStateMutex.RLock()
	defer fake.syncStateMutex.RUnlock()
	return len(fake.syncStateArgsForCall)
}

func (fake *FakeRoom) SyncStateCalls(stub func(types.LocalParticipant, *livekit.SyncState) error) {
	fake.syncStateMutex.Lock()
	defer fake.syncStateMutex.Unlock()
	fake.SyncStateStub = stub
}

func (fake *FakeRoom) SyncStateArgsForCall(i int) (types.LocalParticipant, *livekit.SyncState) {
	fake.syncStateMutex.RLock()
	defer fake.syncStateMutex.RUnlock()
	argsForCall := fake.syncStateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRoom) SyncStateReturns(result1 error) {
	fake.syncStateMutex.Lock()
	defer fake.syncStateMutex.Unlock()
	fake.SyncStateStub = nil
	fake.syncStateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRoom) SyncStateReturnsOnCall(i int, result1 error) {
	fake.syncStateMutex.Lock()
	defer fake.syncStateMutex.Unlock()
	fake.SyncStateStub = nil
	if fake.syncStateReturnsOnCall == nil {
		fake.syncStateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.syncStateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRoom) UpdateParticipantMetadata(arg1 types.LocalParticipant, arg2 string, arg3 string) {
	fake.updateParticipantMetadataMutex.Lock()
	fake.updateParticipantMetadataArgsForCall = append(fake.updateParticipantMetadataArgsForCall, struct {
		arg1 types.LocalParticipant
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	stub := fake.UpdateParticipantMetadataStub
	fake.recordInvocation("UpdateParticipantMetadata", []interface{}{arg1, arg2, arg3})
	fake.updateParticipantMetadataMutex.Unlock()
	if stub != nil {
		fake.UpdateParticipantMetadataStub(arg1, arg2, arg3)
	}
}

func (fake *FakeRoom) UpdateParticipantMetadataCallCount() int {
	fake.updateParticipantMetadataMutex.RLock()
	defer fake.updateParticipantMetadataMutex.RUnlock()
	return len(fake.updateParticipantMetadataArgsForCall)
}

func (fake *FakeRoom) UpdateParticipantMetadataCalls(stub func(types.LocalParticipant, string, string)) {
	fake.updateParticipantMetadataMutex.Lock()
	defer fake.updateParticipantMetadataMutex.Unlock()
	fake.UpdateParticipantMetadataStub = stub
}

func (fake *FakeRoom) UpdateParticipantMetadataArgsForCall(i int) (types.LocalParticipant, string, string) {
	fake.updateParticipantMetadataMutex.RLock()
	defer fake.updateParticipantMetadataMutex.RUnlock()
	argsForCall := fake.updateParticipantMetadataArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeRoom) UpdateSubscriptionPermission(arg1 types.LocalParticipant, arg2 *livekit.SubscriptionPermission) error {
	fake.updateSubscriptionPermissionMutex.Lock()
	ret, specificReturn := fake.updateSubscriptionPermissionReturnsOnCall[len(fake.updateSubscriptionPermissionArgsForCall)]
	fake.updateSubscriptionPermissionArgsForCall = append(fake.updateSubscriptionPermissionArgsForCall, struct {
		arg1 types.LocalParticipant
		arg2 *livekit.SubscriptionPermission
	}{arg1, arg2})
	stub := fake.UpdateSubscriptionPermissionStub
	fakeReturns := fake.updateSubscriptionPermissionReturns
	fake.recordInvocation("UpdateSubscriptionPermission", []interface{}{arg1, arg2})
	fake.updateSubscriptionPermissionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeRoom) UpdateSubscriptionPermissionCallCount() int {
	fake.updateSubscriptionPermissionMutex.RLock()
	defer fake.updateSubscriptionPermissionMutex.RUnlock()
	return len(fake.updateSubscriptionPermissionArgsForCall)
}

func (fake *FakeRoom) UpdateSubscriptionPermissionCalls(stub func(types.LocalParticipant, *livekit.SubscriptionPermission) error) {
	fake.updateSubscriptionPermissionMutex.Lock()
	defer fake.updateSubscriptionPermissionMutex.Unlock()
	fake.UpdateSubscriptionPermissionStub = stub
}

func (fake *FakeRoom) UpdateSubscriptionPermissionArgsForCall(i int) (types.LocalParticipant, *livekit.SubscriptionPermission) {
	fake.updateSubscriptionPermissionMutex.RLock()
	defer fake.updateSubscriptionPermissionMutex.RUnlock()
	argsForCall := fake.updateSubscriptionPermissionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeRoom) UpdateSubscriptionPermissionReturns(result1 error) {
	fake.updateSubscriptionPermissionMutex.Lock()
	defer fake.updateSubscriptionPermissionMutex.Unlock()
	fake.UpdateSubscriptionPermissionStub = nil
	fake.updateSubscriptionPermissionReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeRoom) UpdateSubscriptionPermissionReturnsOnCall(i int, result1 error) {
	fake.updateSubscriptionPermissionMutex.Lock()
	defer fake.updateSubscriptionPermissionMutex.Unlock()
	fake.UpdateSubscriptionPermissionStub = nil
	if fake.updateSubscriptionPermissionReturnsOnCall == nil {
		fake.updateSubscriptionPermissionReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.updateSubscriptionPermissionReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeRoom) UpdateSubscriptions(arg1 types.LocalParticipant, arg2 []livekit.TrackID, arg3 []*livekit.ParticipantTracks, arg4 bool) {
	var arg2Copy []livekit.TrackID
	if arg2 != nil {
		arg2Copy = make([]livekit.TrackID, len(arg2))
		copy(arg2Copy, arg2)
	}
	var arg3Copy []*livekit.ParticipantTracks
	if arg3 != nil {
		arg3Copy = make([]*livekit.ParticipantTracks, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.updateSubscriptionsMutex.Lock()
	fake.updateSubscriptionsArgsForCall = append(fake.updateSubscriptionsArgsForCall, struct {
		arg1 types.LocalParticipant
		arg2 []livekit.TrackID
		arg3 []*livekit.ParticipantTracks
		arg4 bool
	}{arg1, arg2Copy, arg3Copy, arg4})
	stub := fake.UpdateSubscriptionsStub
	fake.recordInvocation("UpdateSubscriptions", []interface{}{arg1, arg2Copy, arg3Copy, arg4})
	fake.updateSubscriptionsMutex.Unlock()
	if stub != nil {
		fake.UpdateSubscriptionsStub(arg1, arg2, arg3, arg4)
	}
}

func (fake *FakeRoom) UpdateSubscriptionsCallCount() int {
	fake.updateSubscriptionsMutex.RLock()
	defer fake.updateSubscriptionsMutex.RUnlock()
	return len(fake.updateSubscriptionsArgsForCall)
}

func (fake *FakeRoom) UpdateSubscriptionsCalls(stub func(types.LocalParticipant, []livekit.TrackID, []*livekit.ParticipantTracks, bool)) {
	fake.updateSubscriptionsMutex.Lock()
	defer fake.updateSubscriptionsMutex.Unlock()
	fake.UpdateSubscriptionsStub = stub
}

func (fake *FakeRoom) UpdateSubscriptionsArgsForCall(i int) (types.LocalParticipant, []livekit.TrackID, []*livekit.ParticipantTracks, bool) {
	fake.updateSubscriptionsMutex.RLock()
	defer fake.updateSubscriptionsMutex.RUnlock()
	argsForCall := fake.updateSubscriptionsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeRoom) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getLocalParticipantsMutex.RLock()
	defer fake.getLocalParticipantsMutex.RUnlock()
	fake.iDMutex.RLock()
	defer fake.iDMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.removeParticipantMutex.RLock()
	defer fake.removeParticipantMutex.RUnlock()
	fake.resolveMediaTrackForSubscriberMutex.RLock()
	defer fake.resolveMediaTrackForSubscriberMutex.RUnlock()
	fake.simulateScenarioMutex.RLock()
	defer fake.simulateScenarioMutex.RUnlock()
	fake.syncStateMutex.RLock()
	defer fake.syncStateMutex.RUnlock()
	fake.updateParticipantMetadataMutex.RLock()
	defer fake.updateParticipantMetadataMutex.RUnlock()
	fake.updateSubscriptionPermissionMutex.RLock()
	defer fake.updateSubscriptionPermissionMutex.RUnlock()
	fake.updateSubscriptionsMutex.RLock()
	defer fake.updateSubscriptionsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRoom) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ types.Room = new(FakeRoom)
