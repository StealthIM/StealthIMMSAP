package nats

import "google.golang.org/protobuf/proto"

// Publish 发布消息（使用 JetStream）
func Publish(subject Subject, data proto.Message) error {
	dataByte, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	conn, err := chooseConn()
	if err != nil {
		return err
	}
	_, err = (*conn.JetStream).Publish(subjectToString[subject], dataByte)
	return err
}

// PublishWithSubject 发布消息（使用 JetStream）
func PublishWithSubject(subject string, data proto.Message) error {
	dataByte, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	conn, err := chooseConn()
	if err != nil {
		return err
	}
	_, err = (*conn.JetStream).Publish(subject, dataByte)
	return err
}

// PublishCore 发布消息（使用 NATS Core）
func PublishCore(subject Subject, data proto.Message) error {
	dataByte, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	conn, err := chooseConn()
	if err != nil {
		return err
	}
	err = (*conn.Conn).Publish(subjectToString[subject], dataByte)
	return err
}
