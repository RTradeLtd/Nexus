package temporal

import "github.com/RTradeLtd/database/models"

// PrivateNetworks is an interface to wrap the Temporal IPFSNetworkManager
// database class
type PrivateNetworks interface {
	GetNetworkByName(name string) (*models.HostedIPFSPrivateNetwork, error)
	UpdateNetworkByName(name string, attrs map[string]interface{}) error
	SaveNetwork(n *models.HostedIPFSPrivateNetwork) error
}
