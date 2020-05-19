package ibmcloud

import (
	"log"
	"time"
)

var endpoints *IdentityEndpoints

func cacheIdentityEndpoints() error {
	if endpoints == nil {
		var err error
		endpoints, err = getIdentityEndpoints()
		if err != nil {
			return err
		}
	}
	return nil
}

func GetIdentityEndpoints() (*IdentityEndpoints, error) {
	err := cacheIdentityEndpoints()
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func GetZones(showFlavors, location string) ([]Zone, error) {
	return getZones(showFlavors, location)
}

func GetVersions() (*ClusterVersion, error) {
	return getVersions()
}

func GetLocations() ([]Location, error) {
	return getLocations()
}

func GetGeoLocations(geo string) ([]Location, error) {
	return getGeoLocations(geo)
}

func GetMachineType(datacenter, serverType, os string, cpuLimit, memoryLimit int) ([]MachineFlavor, error) {
	return getMachineTypes(datacenter, serverType, os, cpuLimit, memoryLimit)
}

func IAMAuthenticate(apikey string) (*Session, error) {
	err := cacheIdentityEndpoints()
	if err != nil {
		log.Println("error with cached data")
		return nil, err
	}
	token, err := getTokenFromIAM(endpoints.TokenEndpoint, apikey)

	if err != nil {
		log.Println("error with token data")
		return nil, err
	}
	return &Session{Token: token}, nil
}

func Authenticate(otp string) (*Session, error) {
	err := cacheIdentityEndpoints()
	if err != nil {
		log.Println("error with cached data")
		return nil, err
	}
	token, err := getToken(endpoints.TokenEndpoint, otp)

	if err != nil {
		log.Println("error with token data")
		return nil, err
	}
	return &Session{Token: token}, nil
}

func (s *Session) GetAccounts() (*Accounts, error) {
	return s.getAccountsWithEndpoint(nil)
}

func (s *Session) IsValid() bool {
	now := time.Now().Unix()
	difference := int64(s.Token.Expiration) - now
	return difference > 100 // expires in 3600 second. keeping 100 second buffer
}

func (s *Session) getAccountsWithEndpoint(nextURL *string) (*Accounts, error) {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, "")
		if err != nil {
			return nil, err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	accounts, err := getAccounts(nextURL, s.Token.AccessToken)
	if err != nil {
		return nil, err
	}
	if accounts.NextURL != nil {
		nextAccounts, err := s.getAccountsWithEndpoint(accounts.NextURL)
		if err != nil {
			return nil, err
		}
		nextAccounts.Resources = append(nextAccounts.Resources, accounts.Resources...)
		return nextAccounts, nil
	}
	return accounts, nil
}

func (s *Session) GetCluster(clusterID, resourceGroup string) (*Cluster, error) {
	return getCluster(s.Token.AccessToken, clusterID, resourceGroup)
}

func (s *Session) GetClusters(location string) ([]*Cluster, error) {
	return getClusters(s.Token.AccessToken, location)
}

func (s *Session) GetDatacenterVlan(datacenter string) ([]Vlan, error) {
	return getDatacenterVlan(s.Token.AccessToken, s.Token.RefreshToken, datacenter)
}

func (s *Session) GetAccountResources(accountID string) (*AccountResources, error) {
	return getAccountResources(s.Token.AccessToken, accountID)
}

func (s *Session) GetWorkers(clusterID string) ([]Worker, error) {
	return getClusterWorkers(s.Token.AccessToken, clusterID)
}

func bindAccountToToken(refreshToken, accountID string) (*Session, error) {
	err := cacheIdentityEndpoints()
	if err != nil {
		return nil, err
	}
	token, err := upgradeToken(endpoints.TokenEndpoint, refreshToken, accountID)
	if err != nil {
		return nil, err
	}
	return &Session{Token: token}, nil
}

func (s *Session) BindAccountToToken(accountID string) (*Session, error) {
	session, err := bindAccountToToken(s.Token.RefreshToken, accountID)
	if err != nil {
		return nil, err
	}
	return session, err
}

func (s *Session) RenewSession() (*Session, error) {
	err := cacheIdentityEndpoints()
	if err != nil {
		return nil, err
	}
	token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, "")
	if err != nil {
		return nil, err
	}
	return &Session{Token: token}, nil
}

func (s *Session) DeleteCluster(id, resourceGroup, deleteResources string) error {
	return deleteCluster(s.Token.AccessToken, id, resourceGroup, deleteResources)
}

func (s *Session) CreateCluster(request CreateClusterRequest) (*CreateClusterResponse, error) {
	return createCluster(s.Token.AccessToken, request)
}

func (s *Session) SetTag(updateTag UpdateTag) (*TagResult, error) {
	return setTags(s.Token.AccessToken, updateTag)
}

func (s *Session) SetClusterTag(tag, clusterID, resourceID string) (*TagResult, error) {
	return setClusterTags(s.Token.AccessToken, tag, clusterID, resourceID)
}

func (s *Session) DeleteTag(updateTag UpdateTag) (*TagResult, error) {
	return deleteTags(s.Token.AccessToken, updateTag)
}

func (s *Session) GetTags(clusterCRN string) (*Tags, error) {
	return getTags(s.Token.AccessToken, clusterCRN)
}

func (s *Session) GetBillingData(accountID, clusterID, clusterCRN string) (string, error) {
	return getBillingData(s.Token.AccessToken, accountID, clusterID, clusterCRN)
}

// CLOUDANT RELATED METHODS
func (s *Session) SetAPIKey(apiKey, accountID string) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return SetAPIKey(apiKey, accountID)
}

func (s *Session) CheckAPIKey(accountID string) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return CheckAPIKey(accountID)
}

func (s *Session) UpdateAPIKey(apiKey, accountID string) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return UpdateAPIKey(apiKey, accountID)
}

func (s *Session) DeleteAPIKey(accountID string) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return DeleteAPIKey(accountID)
}

func (s *Session) GetDocumentV2(accountID string) ([]ScheduleV2, error) {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return nil, err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return GetDocumentV2(accountID)
}

func (s *Session) GetDocument(accountID string) ([]Schedule, error) {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return nil, err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return GetDocument(accountID)
}

func (s *Session) GetAllDocument(accountID string) ([]interface{}, error) {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return nil, err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return GetAllDocument(accountID)
}

func (s *Session) CreateDocument(accountID string, data interface{}) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return CreateDocument(accountID, data)
}

func (s *Session) DeleteDocument(accountID, id, rev string) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}

	return DeleteDocument(accountID, id, rev)
}

func (s *Session) UpdateDocument(accountID, id, rev string, data interface{}) error {
	if !s.IsValid() {
		log.Println("Access token expired.")
		token, err := upgradeToken(endpoints.TokenEndpoint, s.Token.RefreshToken, accountID)
		if err != nil {
			return err
		}
		log.Println("Token Refreshed.")
		s.Token = token
	}
	return UpdateDocument(accountID, id, rev, data)
}