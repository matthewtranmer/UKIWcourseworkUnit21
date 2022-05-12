package signing

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

func send(socket io.Writer, data []byte) error {
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(data)))

	data = append(length, data...)

	_, err := socket.Write(data)
	return err
}

func recv(socket io.Reader) ([]byte, error) {
	buffer := make([]byte, 4)
	_, err := socket.Read(buffer)
	if err != nil {
		return nil, err
	}

	length := binary.LittleEndian.Uint32(buffer)

	buffer = make([]byte, length)
	_, err = socket.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func contactServerJSON(message map[string]string) (map[string]string, error) {
	remote_addr := new(net.TCPAddr)
	remote_addr.IP = []byte{127, 0, 0, 1}
	remote_addr.Port = 50508

	connection, err := net.DialTCP("tcp4", nil, remote_addr)
	if err != nil {
		return nil, err
	}

	json_message, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	err = send(connection, json_message)
	if err != nil {
		return nil, err
	}

	json_response, err := recv(connection)
	if err != nil {
		return nil, err
	}

	response := *new(map[string]string)
	err = json.Unmarshal(json_response, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func GenerateSignature(payload string) (signature, public_key string, err error) {
	req := map[string]string{"command": "generate", "payload": payload}
	response, err := contactServerJSON(req)
	if err != nil {
		return "", "", err
	}

	return response["signature"], response["public key"], nil
}

func VerifySignature(payload, signature, public_key string) (bool, error) {
	req := map[string]string{"command": "verify", "payload": payload, "signature": signature, "public key": public_key}
	response, err := contactServerJSON(req)
	if err != nil {
		return false, err
	}

	return response["is valid"] == "True", nil
}

func BlacklistSignature(payload, signature, public_key string) (bool, error) {
	req := map[string]string{"command": "blacklist", "payload": payload, "signature": signature, "public key": public_key}
	response, err := contactServerJSON(req)
	if err != nil {
		return false, err
	}

	return response["success"] == "True", nil
}
