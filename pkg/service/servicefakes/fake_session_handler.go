// Code generated by counterfeiter. DO NOT EDIT.
package servicefakes

import (
	"context"
	"sync"

	"github.com/livekit/livekit-server/pkg/routing"
	"github.com/livekit/livekit-server/pkg/service"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
)

type FakeSessionHandler struct {
	HandleSessionStub        func(context.Context, livekit.RoomName, routing.ParticipantInit, livekit.ConnectionID, routing.MessageSource, routing.MessageSink) error
	handleSessionMutex       sync.RWMutex
	handleSessionArgsForCall []struct {
		arg1 context.Context
		arg2 livekit.RoomName
		arg3 routing.ParticipantInit
		arg4 livekit.ConnectionID
		arg5 routing.MessageSource
		arg6 routing.MessageSink
	}
	handleSessionReturns struct {
		result1 error
	}
	handleSessionReturnsOnCall map[int]struct {
		result1 error
	}
	LoggerStub        func(context.Context) logger.Logger
	loggerMutex       sync.RWMutex
	loggerArgsForCall []struct {
		arg1 context.Context
	}
	loggerReturns struct {
		result1 logger.Logger
	}
	loggerReturnsOnCall map[int]struct {
		result1 logger.Logger
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSessionHandler) HandleSession(arg1 context.Context, arg2 livekit.RoomName, arg3 routing.ParticipantInit, arg4 livekit.ConnectionID, arg5 routing.MessageSource, arg6 routing.MessageSink) error {
	fake.handleSessionMutex.Lock()
	ret, specificReturn := fake.handleSessionReturnsOnCall[len(fake.handleSessionArgsForCall)]
	fake.handleSessionArgsForCall = append(fake.handleSessionArgsForCall, struct {
		arg1 context.Context
		arg2 livekit.RoomName
		arg3 routing.ParticipantInit
		arg4 livekit.ConnectionID
		arg5 routing.MessageSource
		arg6 routing.MessageSink
	}{arg1, arg2, arg3, arg4, arg5, arg6})
	stub := fake.HandleSessionStub
	fakeReturns := fake.handleSessionReturns
	fake.recordInvocation("HandleSession", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6})
	fake.handleSessionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4, arg5, arg6)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeSessionHandler) HandleSessionCallCount() int {
	fake.handleSessionMutex.RLock()
	defer fake.handleSessionMutex.RUnlock()
	return len(fake.handleSessionArgsForCall)
}

func (fake *FakeSessionHandler) HandleSessionCalls(stub func(context.Context, livekit.RoomName, routing.ParticipantInit, livekit.ConnectionID, routing.MessageSource, routing.MessageSink) error) {
	fake.handleSessionMutex.Lock()
	defer fake.handleSessionMutex.Unlock()
	fake.HandleSessionStub = stub
}

func (fake *FakeSessionHandler) HandleSessionArgsForCall(i int) (context.Context, livekit.RoomName, routing.ParticipantInit, livekit.ConnectionID, routing.MessageSource, routing.MessageSink) {
	fake.handleSessionMutex.RLock()
	defer fake.handleSessionMutex.RUnlock()
	argsForCall := fake.handleSessionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6
}

func (fake *FakeSessionHandler) HandleSessionReturns(result1 error) {
	fake.handleSessionMutex.Lock()
	defer fake.handleSessionMutex.Unlock()
	fake.HandleSessionStub = nil
	fake.handleSessionReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeSessionHandler) HandleSessionReturnsOnCall(i int, result1 error) {
	fake.handleSessionMutex.Lock()
	defer fake.handleSessionMutex.Unlock()
	fake.HandleSessionStub = nil
	if fake.handleSessionReturnsOnCall == nil {
		fake.handleSessionReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.handleSessionReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeSessionHandler) Logger(arg1 context.Context) logger.Logger {
	fake.loggerMutex.Lock()
	ret, specificReturn := fake.loggerReturnsOnCall[len(fake.loggerArgsForCall)]
	fake.loggerArgsForCall = append(fake.loggerArgsForCall, struct {
		arg1 context.Context
	}{arg1})
	stub := fake.LoggerStub
	fakeReturns := fake.loggerReturns
	fake.recordInvocation("Logger", []interface{}{arg1})
	fake.loggerMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeSessionHandler) LoggerCallCount() int {
	fake.loggerMutex.RLock()
	defer fake.loggerMutex.RUnlock()
	return len(fake.loggerArgsForCall)
}

func (fake *FakeSessionHandler) LoggerCalls(stub func(context.Context) logger.Logger) {
	fake.loggerMutex.Lock()
	defer fake.loggerMutex.Unlock()
	fake.LoggerStub = stub
}

func (fake *FakeSessionHandler) LoggerArgsForCall(i int) context.Context {
	fake.loggerMutex.RLock()
	defer fake.loggerMutex.RUnlock()
	argsForCall := fake.loggerArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeSessionHandler) LoggerReturns(result1 logger.Logger) {
	fake.loggerMutex.Lock()
	defer fake.loggerMutex.Unlock()
	fake.LoggerStub = nil
	fake.loggerReturns = struct {
		result1 logger.Logger
	}{result1}
}

func (fake *FakeSessionHandler) LoggerReturnsOnCall(i int, result1 logger.Logger) {
	fake.loggerMutex.Lock()
	defer fake.loggerMutex.Unlock()
	fake.LoggerStub = nil
	if fake.loggerReturnsOnCall == nil {
		fake.loggerReturnsOnCall = make(map[int]struct {
			result1 logger.Logger
		})
	}
	fake.loggerReturnsOnCall[i] = struct {
		result1 logger.Logger
	}{result1}
}

func (fake *FakeSessionHandler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.handleSessionMutex.RLock()
	defer fake.handleSessionMutex.RUnlock()
	fake.loggerMutex.RLock()
	defer fake.loggerMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSessionHandler) recordInvocation(key string, args []interface{}) {
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

var _ service.SessionHandler = new(FakeSessionHandler)
