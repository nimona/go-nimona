package node

// Address returns the node's address
func (n *Node) Address() string {
	return n.container.Address
}
