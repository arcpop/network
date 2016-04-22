package shell

import (
	"fmt"
	"github.com/arcpop/network/netdev"
	"net"
	"github.com/arcpop/network/ipv4"
)

var ifaceHelp = "iface - Possible commands:\n" + 
    "\tiface -> Prints info on all interfaces\n" + 
    "\tiface <interface> -> Prints info on specified interface\n" +
    "\tiface <interface> add [CIDR] -> Adds the specified interface\n" +
    "\tiface <interface> addr <CIDR> -> Sets address on specified interface,\n\t\taddress should be in CIDR notation\n"

func runIface(args []string) {
    if len(args) < 1 {
        fmt.Println(netdev.GetAllInterfaceInfo())
    } else if args[0] == "help" {
        fmt.Println(ifaceHelp)
    } else if len(args) == 1 { 
        ifacename := args[0]
        iface := netdev.InterfaceByName(ifacename)
        if iface == nil {
            fmt.Println("No interface with name " + ifacename + "found!")
            return
        }
        fmt.Println(netdev.GetInterfaceInfo(iface))
    } else if len(args) == 3 { 
        ifacename := args[0]
        what := args[1]
        cidr := args[2]
        if what != "addr" {
            fmt.Println(ifaceHelp)
            return
        }
        ip, n, err := net.ParseCIDR(cidr)
        if err != nil {
            fmt.Println("Failed to parse address: ", err)
            return
        }
        ip4 := ip.To4()
        if ip4 != nil {
            ipv4.ConfigureInterfaceAddress(ifacename, net.IPNet{IP:ip4, Mask: n.Mask})
        } else {
            //Configure ipv6
        }
    } else {
        fmt.Println(ifaceHelp)
    }
}