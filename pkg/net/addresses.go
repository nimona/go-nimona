package net

type Addresses map[string]bool

func NewAddresses() Addresses {
	return Addresses{}
}

func (a Addresses) Add(addresses ...string) {
	for _, address := range addresses {
		if blacklisted, ok := a[address]; ok || blacklisted {
			continue
		}
		a[address] = true
	}
}

func (a Addresses) Remove(addresses ...string) {
	for _, address := range addresses {
		delete(a, address)
	}
}

func (a Addresses) Blacklist(addresses ...string) {
	for _, address := range addresses {
		a[address] = false
	}
}

func (a Addresses) List() []string {
	addresses := []string{}
	for address, ok := range a {
		if ok {
			addresses = append(addresses, address)
		}
	}
	return addresses
}
