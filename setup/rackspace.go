/**
 *  Copyright 2014 Paul Querna
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package setup

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pquerna/distsync/common"
	"github.com/rackspace/gophercloud"
	identAdminRoles "github.com/rackspace/gophercloud/openstack/identity/v2/extensions/admin/roles"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace"
	identityRoles "github.com/rackspace/gophercloud/rackspace/identity/v2/roles"
	"github.com/rackspace/gophercloud/rackspace/identity/v2/tokens"
	identityUsers "github.com/rackspace/gophercloud/rackspace/identity/v2/users"

	"sort"
)

func rackspaceCreateUser(sc *gophercloud.ServiceClient, name string, roles []*identAdminRoles.Role) (string, error) {

	log.WithFields(log.Fields{
		"username": name,
	}).Info("Creating User")

	theEnabledOsStateApiIsNotAwesome := true
	opts := identityUsers.CreateOpts{
		Username: name,
		Enabled:  &theEnabledOsStateApiIsNotAwesome,
		Email:    "",
	}

	user, err := identityUsers.Create(sc, opts).Extract()
	if err != nil {
		return "", err
	}

	for _, r := range roles {
		log.WithFields(log.Fields{
			"user.id":   user.ID,
			"user.name": user.Name,
			"role.id":   r.ID,
			"role.name": r.Name,
		}).Info("Adding role to user")
		err := identityRoles.AddUserRole(sc, user.ID, r.ID).ExtractErr()
		if err != nil {
			return "", err
		}
	}

	log.WithFields(log.Fields{
		"user.name": user.Name,
	}).Info("Creating API Key")

	apiKey, err := identityUsers.ResetAPIKey(sc, user.ID).Extract()
	if err != nil {
		return "", err
	}

	return apiKey.APIKey, nil
}

func rackspacePromptAuth(ui cli.Ui) (gophercloud.AuthOptions, error) {
	var err error
	a := gophercloud.AuthOptions{}
	a.Username, err = ui.Ask("Rackspace Username: ")
	if err != nil {
		return a, err
	}

	a.APIKey, err = ui.Ask("Rackspace API Key: ")
	if err != nil {
		return a, err
	}

	return a, nil
}

func rackspaceAuth(ui cli.Ui) (gophercloud.AuthOptions, error) {
	auth, err := rackspace.AuthOptionsFromEnv()
	if err == nil {
		return auth, nil
	}

	return rackspacePromptAuth(ui)
}

func getRoles(roles []identAdminRoles.Role, rstrs ...string) []*identAdminRoles.Role {
	rv := make([]*identAdminRoles.Role, 0, len(rstrs))
	for _, r := range rstrs {
		// O(n^over9000), but JFDI
		rv = append(rv, getRole(roles, r))
	}
	return rv
}

func getRole(roles []identAdminRoles.Role, roleName string) *identAdminRoles.Role {
	for _, r := range roles {
		if r.Name == roleName {
			return &r
		}
	}

	return nil
}

func getRackspaceRegions(sc *gophercloud.ServiceClient, auth gophercloud.AuthOptions) ([]string, error) {
	regions := make(sort.StringSlice, 0)

	catalog, err := tokens.Create(sc, tokens.WrapOptions(auth)).ExtractServiceCatalog()

	if err != nil {
		return nil, err
	}

	tmp := make(map[string]bool)
	for _, service := range catalog.Entries {
		for _, ep := range service.Endpoints {
			if ep.Region != "" {
				tmp[ep.Region] = true
			}
		}
	}

	for k := range tmp {
		regions = append(regions, k)
	}

	regions.Sort()

	return regions, nil
}

func getRackspaceRoles(sc *gophercloud.ServiceClient) ([]identAdminRoles.Role, error) {
	pager := identityRoles.List(sc)

	allRoles := make([]identAdminRoles.Role, 0)
	err := pager.EachPage(func(p pagination.Page) (bool, error) {
		roles, err := identAdminRoles.ExtractRoles(p)
		if err != nil {
			return false, err
		}
		allRoles = append(allRoles, roles...)
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return allRoles, nil
}

func rackspaceRegion(ui cli.Ui, sc *gophercloud.ServiceClient, auth gophercloud.AuthOptions) (string, error) {
	regions, err := getRackspaceRegions(sc, auth)
	if err != nil {
		return "", err
	}

	choice, err := common.Choice(ui, "Which Rackspace Region should distsync upload files to?", regions)
	if err != nil {
		return "", err
	}

	return regions[choice], nil
}

func Rackspace(ui cli.Ui) (*common.Conf, *common.Conf, error) {
	si, err := newSetupInfo()
	if err != nil {
		return nil, nil, err
	}

	auth, err := rackspaceAuth(ui)
	if err != nil {
		return nil, nil, err
	}

	ac, err := rackspace.AuthenticatedClient(auth)
	if err != nil {
		return nil, nil, err
	}

	sc := rackspace.NewIdentityV2(ac)

	region, err := rackspaceRegion(ui, sc, auth)
	if err != nil {
		return nil, nil, err
	}

	roles, err := getRackspaceRoles(sc)
	if err != nil {
		return nil, nil, err
	}

	uploader := "distsyncUpload" + si.Id
	downloader := "distsyncDownload" + si.Id

	keyUploader, err := rackspaceCreateUser(sc, uploader, getRoles(roles, "object-store:admin"))
	if err != nil {
		return nil, nil, err
	}

	clientconf := common.NewConf()
	clientconf.SharedSecret = si.SharedSecret
	clientconf.StorageBucket = si.BucketName
	clientconf.Rackspace = &common.RackspaceCreds{
		Region:   region,
		Username: uploader,
		ApiKey:   keyUploader,
	}

	keyDownloader, err := rackspaceCreateUser(sc, downloader, getRoles(roles, "object-store:observer"))
	if err != nil {
		return nil, nil, err
	}

	serverconf := common.NewConf()
	clientconf.SharedSecret = si.SharedSecret
	clientconf.StorageBucket = si.BucketName
	serverconf.Rackspace = &common.RackspaceCreds{
		Region:   region,
		Username: downloader,
		ApiKey:   keyDownloader,
	}

	return clientconf, serverconf, nil
}
