package temporal

import "github.com/RTradeLtd/database/models"

// PrivateNetworks is an interface to wrap the Temporal IPFSNetworkManager
// database class
type PrivateNetworks interface {
	GetNetworkByName(name string) (*models.HostedNetwork, error)
	UpdateNetworkByName(name string, attrs map[string]interface{}) error
	SaveNetwork(n *models.HostedNetwork) error

	GetOfflineNetworks(disabled bool) ([]*models.HostedNetwork, error)
}
