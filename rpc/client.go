package rpc

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	msgpack "github.com/msgpack/msgpack-go"
	"github.com/ugorji/go/codec"
)

type Session struct {
	transport    io.ReadWriteCloser
	autoCoercing bool
	nextId       int
}

func coerce(arguments []interface{}) []interface{} {
	_arguments := make([]interface{}, len(arguments))
	for i, v := range arguments {
		switch _v := v.(type) {
		case string:
			_arguments[i] = []byte(_v)
		default:
			_arguments[i] = _v
		}
	}
	return _arguments
}

// CoerceInt takes a reflected value and returns it as an int64
// panics if not an integer type
func CoerceInt(v reflect.Value) int64 {
	if isIntType(v) {
		return v.Int()
	}

	if isUintType(v) {
		return int64(v.Uint())
	}

	panic("not integer type")
}

// CoerceUint takes a reflected value and returns it as an uint64
// panics if not an integer type
func CoerceUint(v reflect.Value) uint64 {

	if isUintType(v) {
		return v.Uint()
	}

	if isIntType(v) {
		return uint64(v.Int())
	}

	panic("not integer type")
}

// Sends a RPC request to the server.
func (self *Session) SendV(funcName string, arguments []interface{}) (reflect.Value, error) {
	var msgId = self.nextId
	self.nextId += 1
	if self.autoCoercing {
		arguments = coerce(arguments)
	}
	err := SendRequestMessage(self.transport.(io.Writer), msgId, funcName, arguments)
	if err != nil {
		return reflect.Value{}, errors.New("Failed to send a request message: " + err.Error())
	}
	_msgId, result, err := ReceiveResponse(self.transport.(io.Reader))
	if err != nil {
		return reflect.Value{}, err
	}
	if msgId != _msgId {
		return reflect.Value{}, errors.New(fmt.Sprintf("Message IDs don't match (%d != %d)", msgId, _msgId))
	}
	if self.autoCoercing {
		_result := result
		if _result.Kind() == reflect.Array || _result.Kind() == reflect.Slice {
			elemType := _result.Type().Elem()
			if elemType.Kind() == reflect.Uint8 {
				result = reflect.ValueOf(string(_result.Interface().([]byte)))
			}
		}
	}
	return result, nil
}

// Sends a RPC request to the server.
func (self *Session) Send(funcName string, arguments ...interface{}) (reflect.Value, error) {
	return self.SendV(funcName, arguments)
}

// Creates a new session with the specified connection.  Strings are
// automatically converted into raw bytes if autoCoercing is
// enabled.
func NewSession(transport io.ReadWriteCloser, autoCoercing bool) *Session {
	return &Session{transport, autoCoercing, 1}
}

// This is a low-level function that is not supposed to be called directly
// by the user.  Change this if the MessagePack protocol is updated.
func SendRequestMessage(writer io.Writer, msgId int, funcName string, arguments []interface{}) error {
	_, err := writer.Write([]byte{0x94})
	if err != nil {
		return err
	}
	_, err = msgpack.PackInt(writer, REQUEST)
	if err != nil {
		return err
	}
	_, err = msgpack.PackInt(writer, msgId)
	if err != nil {
		return err
	}
	_, err = msgpack.PackBytes(writer, []byte(funcName))
	if err != nil {
		return err
	}
	_, err = msgpack.PackArray(writer, reflect.ValueOf(arguments))
	return err
}

// This is a low-level function that is not supposed to be called directly
// by the user.  Change this if the MessagePack protocol is updated.
func ReceiveResponse(reader io.Reader) (int, reflect.Value, error) {
	msgId, result, err := _ReceiveResponse(reader)
	if err != nil {
		return 0, reflect.Value{}, err
	}
	return int(msgId), reflect.ValueOf(result), nil
}

func _ReceiveResponse(reader io.Reader) (msgId int, result interface{}, err error) {
	var h codec.Handle = new(codec.MsgpackHandle)
	var dec *codec.Decoder = codec.NewDecoder(reader, h)
	var iface []interface{}
	dec.Decode(&iface)
	if len(iface) != 4 {
		return 0, nil, fmt.Errorf("Invalid message. Not enough content")
	}
	if msgType, err := strconv.Atoi(fmt.Sprint(iface[0])); err != nil {
		return 0, nil, fmt.Errorf("Invalid message Type: %T : %v: %v", iface[0], iface[0], err)
	} else {
		if msgType != RESPONSE {
			return 0, nil, fmt.Errorf("Non-Respsonse message Type: %v", msgType)
		}
	}
	if msgId, err = strconv.Atoi(fmt.Sprint(iface[1])); err != nil {
		return 0, nil, fmt.Errorf("Invalid message Id: %T :  %v", iface[1], iface[1])
	}
	if iface[3] != nil {
		err = fmt.Errorf("%v", string(iface[3].([]byte)))
	}
	return msgId, iface[2], err
}
