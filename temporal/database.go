package temporal

import "github.com/RTradeLtd/database/models"

// AccessChecker is an interface to wrap the Temporal UserManager database class
type AccessChecker interface {
	CheckIfUserHasAccessToNetwork(user string, network string) (ok bool, err error)
}

// PrivateNetworks is an interface to wrap the Temporal IPFSNetworkManager
// database class
type PrivateNetworks interface {
	GetNetworkByName(name string) (*models.HostedIPFSPrivateNetwork, error)
	UpdateNetworkByName(name string, attrs map[string]interface{}) error
	SaveNetwork(n *models.HostedIPFSPrivateNetwork) error
}
