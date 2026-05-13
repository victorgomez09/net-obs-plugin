package sniffer

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/victorgomez09/net-obs-plugin/internal/informer"
)

func StartSniffer(iface string, cache *informer.PodCache) {
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Printf("Error opening interface %s: %v", iface, err)
		return
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Process packet and enrich with pod info from cache
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)

			srcPod, _ := cache.Get(ip.SrcIP.String())
			dstPod, _ := cache.Get(ip.DstIP.String())

			if srcPod != "" || dstPod != "" {
				fmt.Printf("[NET] %s (%s) -> %s (%s) | Proto: %s\n",
					srcPod, ip.SrcIP, dstPod, ip.DstIP, ip.Protocol)
			}
		}
	}
}
