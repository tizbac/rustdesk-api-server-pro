package service

import (
	"crypto/tls"
	"fmt"
	"rustdesk-api-server-pro/config"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type LdapService struct {
	config *config.ServerConfig
}

var ldapService *LdapService

func NewLdapService() *LdapService {
	if ldapService != nil {
		return ldapService
	}

	cfg := config.GetServerConfig()
	return &LdapService{
		config: cfg,
	}
}

func (s *LdapService) IsEnabled() bool {
	return s.config.Ldap != nil && s.config.Ldap.Enabled
}

func (s *LdapService) Authenticate(username, password string) (*ldap.Conn, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("LDAP is not enabled")
	}

	ldapConfig := s.config.Ldap
	server := ldapConfig.Host
	port := ldapConfig.Port

	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", server, port))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to LDAP server: %v", err)
	}

	if ldapConfig.UseTLS {
		conn.StartTLS(&tls.Config{
			InsecureSkipVerify: ldapConfig.InsecureSkipVerify,
		})
	}

	if ldapConfig.StartTLS && !ldapConfig.UseTLS {
		err = conn.StartTLS(&tls.Config{
			InsecureSkipVerify: ldapConfig.InsecureSkipVerify,
		})
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("Failed to start TLS: %v", err)
		}
	}

	err = conn.Bind(ldapConfig.BindDN, ldapConfig.BindPassword)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to bind with service account: %v", err)
	}

	userDN := strings.Replace(ldapConfig.UserFilter, "%s", username, 1)

	searchRequest := ldap.NewSearchRequest(
		ldapConfig.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		userDN,
		[]string{"dn", "cn", "mail", "userPrincipalName"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil || len(sr.Entries) != 1 {
		conn.Close()
		if len(sr.Entries) != 1 {
			return nil, fmt.Errorf("User not found or too many entries found")
		}
		return nil, fmt.Errorf("Failed to search directory: %v", err)
	}

	userEntry := sr.Entries[0]
	userDn := userEntry.DN

	err = conn.Bind(userDn, password)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("LDAP authentication failed: %v", err)
	}

	return conn, nil
}

func (s *LdapService) GetUserDN(username string) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("LDAP is not enabled")
	}

	ldapConfig := s.config.Ldap
	server := ldapConfig.Host
	port := ldapConfig.Port

	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", server, port))
	if err != nil {
		return "", fmt.Errorf("Failed to connect to LDAP server: %v", err)
	}
	defer conn.Close()

	if ldapConfig.UseTLS {
		conn.StartTLS(&tls.Config{
			InsecureSkipVerify: ldapConfig.InsecureSkipVerify,
		})
	}

	if ldapConfig.StartTLS && !ldapConfig.UseTLS {
		err = conn.StartTLS(&tls.Config{
			InsecureSkipVerify: ldapConfig.InsecureSkipVerify,
		})
		if err != nil {
			return "", fmt.Errorf("Failed to start TLS: %v", err)
		}
	}

	err = conn.Bind(ldapConfig.BindDN, ldapConfig.BindPassword)
	if err != nil {
		return "", fmt.Errorf("Failed to bind with service account: %v", err)
	}

	userDN := strings.Replace(ldapConfig.UserFilter, "%s", username, 1)

	searchRequest := ldap.NewSearchRequest(
		ldapConfig.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		userDN,
		[]string{"dn", "cn", "mail", "userPrincipalName"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil || len(sr.Entries) != 1 {
		if len(sr.Entries) != 1 {
			return "", fmt.Errorf("User not found or too many entries found")
		}
		return "", fmt.Errorf("Failed to search directory: %v", err)
	}

	return sr.Entries[0].DN, nil
}
