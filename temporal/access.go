package temporal

// AccessChecker is an interface to wrap the Temporal UserManager database class
type AccessChecker interface {
	CheckIfUserHasAccessToNetwork(user string, network string) (ok bool, err error)
}
