package mesh

// // ReusableConn is returned for mutex nets
// type ReusableConn interface {
// 	net.Conn
// 	NewConn() (net.Conn, error)
// }

// // MetaConn provides additional methods to get net metadata
// type MetaConn interface {
// 	net.Conn
// 	NewConn() (net.Conn, error)
// }

// // Read implements the Conn Read method.
// func (c *conn) Read(b []byte) (int, error) {
// 	return c.conn.Read(b)
// }

// // Write implements the Conn Write method.
// func (c *conn) Write(b []byte) (int, error) {
// 	return c.conn.Write(b)
// }

// // Close closes the connection.
// func (c *conn) Close() error {
// 	return c.conn.Close()
// }

// // LocalAddr returns the local network
// // The Addr returned is shared by all invocations of LocalAddr, so
// // do not modify it.
// func (c *conn) LocalAddr() net.Addr {
// 	return c.conn.LocalAddr()
// }

// // RemoteAddr returns the remote network
// // The Addr returned is shared by all invocations of RemoteAddr, so
// // do not modify it.
// func (c *conn) RemoteAddr() net.Addr {
// 	return c.conn.RemoteAddr()
// }

// // SetDeadline implements the Conn SetDeadline method.
// func (c *conn) SetDeadline(t time.Time) error {
// 	return c.conn.SetDeadline(t)
// }

// // SetReadDeadline implements the Conn SetReadDeadline method.
// func (c *conn) SetReadDeadline(t time.Time) error {
// 	return c.conn.SetReadDeadline(t)
// }

// // SetWriteDeadline implements the Conn SetWriteDeadline method.
// func (c *conn) SetWriteDeadline(t time.Time) error {
// 	return c.conn.SetWriteDeadline(t)
// }

// func (c *conn) NewConn() (net.Conn, error) {
// 	if c.server {

// 	}
// 	return nil, nil
// }
