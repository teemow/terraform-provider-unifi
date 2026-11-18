# Reference an existing WireGuard VPN server
resource "unifi_vpn_server" "wireguard" {
  name    = "Spiderhome"
  subnet  = "10.100.0.1/24"

  wireguard = {
    port = 51820
  }
}

# Create WireGuard peers
resource "unifi_wireguard_peer" "laptop" {
  name                = "teemow-laptop"
  wireguard_server_id = unifi_vpn_server.wireguard.id
}

resource "unifi_wireguard_peer" "phone" {
  name                = "teemow-phone"
  wireguard_server_id = unifi_vpn_server.wireguard.id
}

# Peer with explicit enabled flag
resource "unifi_wireguard_peer" "guest" {
  name                = "guest-device"
  wireguard_server_id = unifi_vpn_server.wireguard.id
  enabled             = false
}
