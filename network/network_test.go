package network

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockConn struct {
	mock.Mock
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	copy(b, args.Get(0).([]byte))
	return args.Int(1), args.Error(2)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) Close() error {
	return m.Called().Error(0)
}

func TestConnectAndHandshake(t *testing.T) {
	mockConn := new(MockConn)

	str, _ := hex.DecodeString("f9beb4d976657273696f6e000000000064000000358d493262ea0000010000000000000011b2d05000000000010000000000000000000000000000000000ffff000000000000000000000000000000000000000000000000ffff0000000000003b2eb35d8ce617650f2f5361746f7368693a302e372e322fc03e0300")
	strVerack, _ := hex.DecodeString("F9BEB4D976657261636B000000000000000000005DF6E0E2")
	// Define what the mock should return on each read.
	mockConn.On("Read", mock.Anything).Return(str, 24, nil).Twice()
	mockConn.On("Read", mock.Anything).Return(strVerack, 4, nil)
	mockConn.On("Write", mock.Anything).Return(56, nil).Once()
	mockConn.On("Close").Return(nil).Once()

	go func() {
		ConnectAndHandshake(mockConn)
	}()

	time.Sleep(time.Millisecond * 100)

	mockConn.AssertExpectations(t)
}

func TestReadMessages(t *testing.T) {
	mockConn := new(MockConn)
	doneChannel := make(chan struct{})
	receiveChannel := make(chan []byte, 1)
	header := []byte{0xf9, 0xbe, 0xb4, 0xd9, 0x76, 0x65, 0x72, 0x61, 0x63, 0x6b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5D, 0xF6, 0xE0, 0xE2}
	payload := []byte{}

	mockConn.On("Read", mock.Anything).Return(header, len(header), nil).Once()
	mockConn.On("Read", mock.Anything).Return(payload, len(payload), nil)

	go readMessages(context.TODO(), mockConn, receiveChannel)
	time.Sleep(time.Millisecond * 50)

	close(doneChannel)
	assert.Equal(t, append(header, payload...), <-receiveChannel)
	mockConn.AssertExpectations(t)
}

func TestSendMessages(t *testing.T) {
	mockConn := new(MockConn)
	doneChannel := make(chan struct{})
	sendChannel := make(chan Message, 1)

	message := Message{Command: "verack", Payload: []byte{}}
	mockConn.On("Write", mock.Anything).Return(24, nil).Once()

	verackSent := false
	go sendMessages(context.TODO(), mockConn, sendChannel, &verackSent)
	sendChannel <- message
	time.Sleep(time.Millisecond * 50)

	close(doneChannel)
	assert.True(t, verackSent)
	mockConn.AssertExpectations(t)
}
