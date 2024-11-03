// sync_command.go
package main

import (
    "context"
    "fmt"
    "strings"

    "code.gitea.io/sdk/gitea"
    "github.com/Nerzal/gocloak/v13"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync users from Keycloak to Gitea",
		Long:  `Synchronize users from a Keycloak realm to Gitea, creating users and repositories as needed`,
		Run:   runSync,
	}

	// Command flags
	keycloakURL      string
	keycloakRealm    string
	keycloakUser     string
	keycloakPassword string
	giteaURL         string
	giteaToken       string
	giteaOAuthID     int64
	orgName          string
	teamName         string
)

func init() {
	rootCmd.AddCommand(syncCmd)

	// Add flags for the sync command
	syncCmd.Flags().StringVar(&keycloakURL, "keycloak-url", "", "Keycloak server URL")
	syncCmd.Flags().StringVar(&keycloakRealm, "keycloak-realm", "", "Keycloak realm name")
	syncCmd.Flags().StringVar(&keycloakUser, "keycloak-user", "", "Keycloak admin username")
	syncCmd.Flags().StringVar(&keycloakPassword, "keycloak-password", "", "Keycloak admin password")
	syncCmd.Flags().StringVar(&giteaURL, "gitea-url", "", "Gitea server URL")
	syncCmd.Flags().StringVar(&giteaToken, "gitea-token", "", "Gitea admin access token")
	syncCmd.Flags().Int64Var(&giteaOAuthID, "gitea-oauth-id", 0, "Gitea OAuth source ID")
	syncCmd.Flags().StringVar(&orgName, "org-name", "", "Gitea organization name")
	syncCmd.Flags().StringVar(&teamName, "team-name", "", "Gitea team name")

	// Mark required flags
	syncCmd.MarkFlagRequired("keycloak-url")
	syncCmd.MarkFlagRequired("keycloak-realm")
	syncCmd.MarkFlagRequired("keycloak-user")
	syncCmd.MarkFlagRequired("keycloak-password")
	syncCmd.MarkFlagRequired("gitea-url")
	syncCmd.MarkFlagRequired("gitea-token")
	syncCmd.MarkFlagRequired("gitea-oauth-id")
	syncCmd.MarkFlagRequired("org-name")
	syncCmd.MarkFlagRequired("team-name")
}

// KeycloakClient wraps Keycloak operations
type KeycloakClient struct {
	client  *gocloak.GoCloak
	token   *gocloak.JWT
	realm   string
	context context.Context
}

// NewKeycloakClient creates a new Keycloak client
func NewKeycloakClient(url, realm, username, password string) (*KeycloakClient, error) {
	client := gocloak.NewClient(url)
	ctx := context.Background()

	token, err := client.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		return nil, err
	}

	return &KeycloakClient{
		client:  client,
		token:   token,
		realm:   realm,
		context: ctx,
	}, nil
}

// GetUsers retrieves all users from Keycloak
func (k *KeycloakClient) GetUsers() ([]*gocloak.User, error) {
	firstUser := 0
	maxUser := 1024
	return k.client.GetUsers(k.context, k.token.AccessToken, k.realm, gocloak.GetUsersParams{
		First: &firstUser,
		Max:   &maxUser,
	})
}

// GiteaSync handles Gitea synchronization operations
type GiteaSync struct {
	client   *gitea.Client
	oauthID  int64
	orgName  string
	teamName string
	teamID   int64
	orgID    int64
}

// NewGiteaSync creates a new Gitea sync client
func NewGiteaSync(url, token string, oauthID int64, orgName, teamName string) (*GiteaSync, error) {
	client, err := gitea.NewClient(url, gitea.SetToken(token))
	if err != nil {
		return nil, err
	}

	gs := &GiteaSync{
		client:   client,
		oauthID:  oauthID,
		orgName:  orgName,
		teamName: teamName,
	}

	// Initialize org and team IDs
	if err := gs.initOrgAndTeam(); err != nil {
		return nil, err
	}

	return gs, nil
}

// initOrgAndTeam initializes organization and team IDs
func (g *GiteaSync) initOrgAndTeam() error {
	org, _, err := g.client.GetOrg(g.orgName)
	if err != nil {
		return err
	}
	g.orgID = org.ID

	teams, _, err := g.client.ListOrgTeams(g.orgName, gitea.ListTeamsOptions{})
	if err != nil {
		return err
	}

	for _, team := range teams {
		if team.Name == g.teamName {
			g.teamID = team.ID
			return nil
		}
	}

	return fmt.Errorf("team %s not found in organization %s", g.teamName, g.orgName)
}

// SyncUser synchronizes a Keycloak user to Gitea
func (g *GiteaSync) SyncUser(keycloakUser *gocloak.User) error {
	email := *keycloakUser.Email
	repoName := strings.Split(email, "@")[0]
	username := repoName

	// Check if user exists
	_, resp, err := g.client.GetUserInfo(username)
	if err != nil && resp.StatusCode != 404 {
		return err
	}

	if resp.StatusCode == 404 {
		// Create user if not exists
		options := gitea.CreateUserOption{
			Username:  username,
			Email:     email,
			FullName:  *keycloakUser.FirstName + " " + *keycloakUser.LastName,
			LoginName: *keycloakUser.ID,
			SourceID:  g.oauthID,
		}

		_, _, err = g.client.AdminCreateUser(options)
		if err != nil {
			return err
		}
	}

	// Add user to team
	_, err = g.client.AddTeamMember(g.teamID, username)
	if err != nil {
		log.Warnf("Failed to add user %s to team: %v", username, err)
	}

	// Create repository
	_, _, err = g.client.CreateOrgRepo(g.orgName, gitea.CreateRepoOption{
		Name:          repoName,
		Private:       true,
		AutoInit:      true,
		DefaultBranch: "main",
		Description:   "Personal repository for " + username,
	})
	if err != nil {
		log.Errorf("Failed to create repository for user %s: %v", username, err)
		return err
	}

	perm := gitea.AccessModeWrite
	_, err = g.client.AddCollaborator(g.orgName, repoName, username, gitea.AddCollaboratorOption{
		Permission: &perm,
	})
	if err != nil {
		log.Errorf("Failed to add collaborator to repository for user %s: %v", username, err)
		return err
	}

	return nil
}

func runSync(cmd *cobra.Command, args []string) {
	// Initialize Keycloak client
	keycloak, err := NewKeycloakClient(keycloakURL, keycloakRealm, keycloakUser, keycloakPassword)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize Keycloak client")
	}

	// Initialize Gitea sync
	giteClient, err := NewGiteaSync(giteaURL, giteaToken, giteaOAuthID, orgName, teamName)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize Gitea client")
	}

	// Get users from Keycloak
	users, err := keycloak.GetUsers()
	if err != nil {
		log.WithError(err).Fatal("Failed to get users from Keycloak")
	}

	// Sync each user
	for _, user := range users {
		if err := giteClient.SyncUser(user); err != nil {
			log.WithError(err).WithField("username", *user.Username).Error("Failed to sync user")
			continue
		}
		log.WithField("username", *user.Username).Info("Successfully synced user")
	}

	log.Info("Sync completed")
}
