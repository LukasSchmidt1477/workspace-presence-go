package main

import "fmt"

type Account struct {
	ID, Name string
	Active   bool
}
type WorkspaceService struct {
	client   *Client
	accounts map[string]Account
}

func NewWorkspaceService(client *Client) *WorkspaceService {
	return &WorkspaceService{client: client, accounts: map[string]Account{}}
}

func (s *WorkspaceService) Onboard(tenant, accountID, name string) error {
	if tenant == "" || accountID == "" || name == "" {
		return fmt.Errorf("tenant, accountID, and name are required")
	}
	if err := s.client.CreateChannel(tenant); err != nil {
		return err
	}
	s.accounts[accountID] = Account{ID: accountID, Name: name, Active: true}
	return s.client.Publish(tenant, accountID, "account.online", map[string]string{"name": name})
}

func (s *WorkspaceService) SetActive(accountID string, active bool) error {
	a, ok := s.accounts[accountID]
	if !ok {
		return fmt.Errorf("account not found")
	}
	a.Active = active
	s.accounts[accountID] = a
	return nil
}

func (s *WorkspaceService) Online(tenant string) ([]string, error) { return s.client.Presence(tenant) }

func (s *WorkspaceService) PublishIfActive(tenant, accountID string) error {
	a, ok := s.accounts[accountID]
	if !ok || !shouldPublish(a) {
		return fmt.Errorf("account is inactive")
	}
	return s.client.Publish(tenant, accountID, "account.online", map[string]string{"name": a.Name})
}

func shouldPublish(a Account) bool { return a.ID != "" && a.Active }
