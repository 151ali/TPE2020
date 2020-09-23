package main

/* notes :
$ go build netw.go
$ sudo chown root:root netw
$ sudo chmod u+s netw
*/
import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/vishvananda/netlink"
)

const (
	bridgeName = "brdemo0"
	ipAdrr     = "10.1.1.1/24"
)

func main() {
	if len(os.Args) > 1 {
		p, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		pid := p
		fmt.Println(pid)

		/// network stuff
		if err := createBridge(); err != nil {
			fmt.Println("bridge :", err)
		}

		if err := createVethpair(pid); err != nil {
			fmt.Println("pair :", err)
		}
		//
		//
		//
	} else {
		fmt.Println("you should pass a PID as an argument !")
	}
}

//*******************************************************************

// ===========================================================================
// createBridge :
func createBridge() error {
	//check if the bridge is already exist
	_, err := netlink.LinkByName(bridgeName)
	if err == nil {
		return nil
	}

	//create the bridge :
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	mybridge := &netlink.Bridge{LinkAttrs: la}

	if err = netlink.LinkAdd(mybridge); err != nil {
		return fmt.Errorf("Can't create the bridge %s : %v", la.Name, err)
	}

	// add ip adress to "bridgeName"
	addr, err := netlink.ParseAddr(ipAdrr)
	if err != nil {
		return fmt.Errorf("cant parse ip adresse : %v", err)
	}

	if err = netlink.AddrAdd(mybridge, addr); err != nil {
		return fmt.Errorf("can't add addr to bridge : %v", err)
	}

	// set up the bridge :
	// similat to : ip link set dev container0 up
	err = netlink.LinkSetUp(mybridge)
	if err != nil {
		return fmt.Errorf("Can't set up bridge - %s - : %v", bridgeName, err)
	}

	return nil
}

// createVethpair :
func createVethpair(pid int) error {
	// The  veth  devices  are virtual Ethernet devices.  They can act as tunnels between network
	// namespaces to create a bridge to a physical network device in another namespace,  but  can
	// also be used as standalone network devices.

	//==============================================================================
	// attach one side of the veth-pair to the bridge :
	// check if the bridge exist ?
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	//fmt.Println("br", br)

	//gives names for veth pair
	vHost := "veth-host"
	vContainer := "veth-container"

	// create the Veth pair
	la := netlink.NewLinkAttrs()
	la.Name = vHost
	la.MasterIndex = br.Attrs().Index

	// *netlink.Veth
	vp := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  vContainer,
	}
	if err = netlink.LinkAdd(vp); err != nil {
		return fmt.Errorf("error while creating veth pair : %v", err)
	}

	//
	//
	//
	// vContainer as peer
	peer, err := netlink.LinkByName(vContainer)
	if err != nil {
		return fmt.Errorf("can't get peer(vContainer) : %v", err)
	}
	// put peer to the container specified by PID
	if err = netlink.LinkSetNsPid(peer, pid); err != nil {
		return fmt.Errorf("Can't put the peer into network namespace : %v", err)
	}

	// sumup
	err = netlink.LinkSetUp(vp)
	if err != nil {
		return err
	}

	return nil
}
