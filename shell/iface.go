package shell

import (
	"fmt"
	"github.com/arcpop/network/netdev"
	"net"
	"strings"
	"strconv"
)

var ifaceHelp = "iface - Possible commands:\n" + 
    "\tiface -> Prints info on all interfaces\n" + 
    "\tiface <interface> -> Prints info on specified interface\n" +
    "\tiface <interface> <CIDR> -> Sets addresse on specified interface,\n\t\t address should be in CIDR notation\n"

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
    } else if len(args) == 2 { 
        ifacename := args[0]
        cidr := args[1]
        iface := netdev.InterfaceByName(ifacename)
        if iface == nil {
            fmt.Println("No interface with name " + ifacename + "found!")
            return
        }
        ip, n, err := net.ParseCIDR(cidr)
        if err != nil {
            fmt.Println("Failed to parse address: ", err)
            return
        }
        ip4 := ip.To4()
        if ip4 != nil {
            iface.SetIPv4Address(ip4, net.IP(n.Mask).To4())
        } else {
            strs := strings.Split(cidr, "/")
            if len(strs) < 2 {
                fmt.Println("Failed to parse address")
                return
            }
            nm, err := strconv.Atoi(strs[1])
            if err != nil {
                fmt.Println("Failed to parse address: ", err)
                return
            }
            iface.SetIPv6Address(ip, nm)
        }
    } else {
        fmt.Println(ifaceHelp)
    }
}