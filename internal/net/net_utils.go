package net

// func Write(o *object.Object, conn *Connection) error {
// 	if conn == nil {
// 		log.DefaultLogger.Info("conn cannot be nil")
// 		return fmt.Errorf("missing conn")
// 	}

// 	ra := ""
// 	if !conn.RemotePeerKey.IsEmpty() {
// 		ra = conn.RemotePeerKey.String()
// 	}

// 	log.DefaultLogger.Debug(
// 		"writing to connection",
// 		log.Any("object", o),
// 		log.String("local.address", conn.localAddress),
// 		log.String("remote.address", conn.remoteAddress),
// 		log.String("remote.fingerprint", ra),
// 		log.String("direction", "outgoing"),
// 	)

// 	if err := conn.encoder.Encode(o); err != nil {
// 		return fmt.Errorf("error marshaling object: %w", err)
// 	}

// 	return nil
// }

// func Read(conn *Connection) (*object.Object, error) {
// 	logger := log.DefaultLogger

// 	defer func() {
// 		if r := recover(); r != nil {
// 			logger.Error(
// 				"recovered from panic during read",
// 				log.Any("r", r),
// 				log.Stack(),
// 			)
// 		}
// 	}()

// 	o := &object.Object{}
// 	err := conn.decoder.Decode(o)
// 	if err != nil {
// 		return nil, err
// 	}

// 	logger.Debug(
// 		"reading from connection",
// 		log.Any("object", o),
// 		log.String("local.address", conn.localAddress),
// 		log.String("remote.address", conn.remoteAddress),
// 		log.String("remote.publicKey", conn.RemotePeerKey.String()),
// 		log.String("direction", "incoming"),
// 	)

// 	return o, nil
// }
