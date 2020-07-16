// This is lease support for nclient4

package nclient4

import (
	"fmt"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
)

// Lease contains a DHCPv4 lease after DORA.
// note: Lease doesn't include binding interface name
type Lease struct {
	Offer        *dhcpv4.DHCPv4
	ACK          *dhcpv4.DHCPv4
	CreationTime time.Time
}

// NewReleaseFromLease creates a DHCPv4 Release message from the lease.
// default Release message without any Modifer is created as following:
//  - option Message Type is Release
//  - ClientIP is set to lease.ACK.YourIPAddr
//  - ClientHWAddr is set to lease.ACK.ClientHWAddr
//  - Unicast
//  - option Server Identifier is set to ServerIdentifier of lease.ACK
func NewReleaseFromLease(lease *Lease, modifiers ...dhcpv4.Modifier) (*dhcpv4.DHCPv4, error) {
	return dhcpv4.New(dhcpv4.PrependModifiers(modifiers,
		dhcpv4.WithMessageType(dhcpv4.MessageTypeRelease),
		dhcpv4.WithClientIP(lease.ACK.YourIPAddr),
		dhcpv4.WithHwAddr(lease.ACK.ClientHWAddr),
		dhcpv4.WithBroadcast(false),
		dhcpv4.WithOption(dhcpv4.OptServerIdentifier(lease.ACK.ServerIdentifier())),
	)...)
}

// Release send DHCPv4 release messsage to server, based on specified lease.
// release is sent as unicast per RFC2131, section 4.4.4.
// Note: some DHCP server requries of using assigned IP address as source IP,
// use nclient4.WithUnicast to create client for such case.
func (c *Client) Release(lease *Lease, modifiers ...dhcpv4.Modifier) error {
	if lease == nil {
		return fmt.Errorf("lease is nil")
	}
	req, err := NewReleaseFromLease(lease, modifiers...)
	if err != nil {
		return fmt.Errorf("fail to create release request,%w", err)
	}
	_, err = c.conn.WriteTo(req.ToBytes(), &net.UDPAddr{IP: lease.ACK.Options.Get(dhcpv4.OptionServerIdentifier), Port: ServerPort})
	if err == nil {
		c.logger.PrintMessage("sent message:", req)
	}
	return err
}
