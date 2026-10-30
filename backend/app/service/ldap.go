package service

import (
	"crypto/tls"
	"fmt"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

type LdapService struct{}

func NewLdapService() *LdapService {
	return &LdapService{}
}

// LdapConfig returns the yaml defaults merged with the values saved in the panel.
func LdapConfig() config.LdapConfig {
	settings, err := GetRuntimeSettings(db.DbEngine)
	if err != nil {
		cfg := config.GetServerConfig()
		if cfg.Ldap != nil {
			return *cfg.Ldap
		}
		return config.LdapConfig{}
	}

	return settings.Ldap
}

func (s *LdapService) IsEnabled() bool {
	return LdapConfig().Enabled
}

func (s *LdapService) Authenticate(username, password string) (*ldap.Conn, error) {
	ldapConfig := LdapConfig()
	if !ldapConfig.Enabled {
		return nil, fmt.Errorf("LDAP is not enabled")
	}

	conn, err := connectLdap(ldapConfig)
	if err != nil {
		return nil, err
	}

	if err := bindServiceAccount(conn, ldapConfig); err != nil {
		conn.Close()
		return nil, err
	}

	userDN, err := searchUserDN(conn, ldapConfig, username)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := conn.Bind(userDN, password); err != nil {
		conn.Close()
		return nil, fmt.Errorf("LDAP authentication failed: %v", err)
	}

	return conn, nil
}

func (s *LdapService) GetUserDN(username string) (string, error) {
	ldapConfig := LdapConfig()
	if !ldapConfig.Enabled {
		return "", fmt.Errorf("LDAP is not enabled")
	}

	conn, err := connectLdap(ldapConfig)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if err := bindServiceAccount(conn, ldapConfig); err != nil {
		return "", err
	}

	return searchUserDN(conn, ldapConfig, username)
}

// connectLdap opens the connection, TLS is used when use_tls (LDAPS) or
// start_tls is set on a plain connection.
func connectLdap(ldapConfig config.LdapConfig) (*ldap.Conn, error) {
	addr := fmt.Sprintf("%s:%d", ldapConfig.Host, ldapConfig.Port)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: ldapConfig.InsecureSkipVerify,
	}

	if ldapConfig.UseTLS {
		conn, err := ldap.DialTLS("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to LDAP server over TLS: %v", err)
		}
		return conn, nil
	}

	conn, err := ldap.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %v", err)
	}

	if ldapConfig.StartTLS {
		if err := conn.StartTLS(tlsConfig); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to start TLS: %v", err)
		}
	}

	return conn, nil
}

// bindServiceAccount binds the service account. With Active Directory a bare
// account name (no comma) is also tried as an UPN built from the base DN.
func bindServiceAccount(conn *ldap.Conn, ldapConfig config.LdapConfig) error {
	bindDN := strings.TrimSpace(ldapConfig.BindDN)

	if bindDN == "" {
		return fmt.Errorf("no bind DN configured")
	}

	bindErr := conn.Bind(bindDN, ldapConfig.BindPassword)
	if bindErr == nil {
		return nil
	}

	if ldapConfig.ActiveDirectory && !strings.Contains(bindDN, ",") && ldapConfig.BaseDN != "" {
		upn := bindDN + "@" + domainFromBaseDN(ldapConfig.BaseDN)
		if err := conn.Bind(upn, ldapConfig.BindPassword); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to bind with the service account: %v", bindErr)
}

// searchUserDN finds the user with the configured filter and returns its DN.
func searchUserDN(conn *ldap.Conn, ldapConfig config.LdapConfig, username string) (string, error) {
	filter := strings.Replace(ldapConfig.UserFilter, "%s", username, 1)

	searchRequest := ldap.NewSearchRequest(
		ldapConfig.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"dn", "cn", "mail", "userPrincipalName"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return "", fmt.Errorf("failed to search the directory: %v", err)
	}

	if len(sr.Entries) != 1 {
		return "", fmt.Errorf("user not found or too many entries found")
	}

	return sr.Entries[0].DN, nil
}

func domainFromBaseDN(baseDN string) string {
	parts := []string{}
	for _, component := range strings.Split(baseDN, ",") {
		component = strings.TrimSpace(component)
		if strings.HasPrefix(strings.ToLower(component), "dc=") {
			parts = append(parts, component[3:])
		}
	}

	return strings.Join(parts, ".")
}

// LdapTestRequest is the payload of the bind test, it carries the values of the
// form so the connection can be verified before saving them.
type LdapTestRequest struct {
	Enabled            bool   `json:"enabled"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	UseTLS             bool   `json:"use_tls"`
	StartTLS           bool   `json:"start_tls"`
	BindDN             string `json:"bind_dn"`
	BindPassword       string `json:"bind_password"`
	BaseDN             string `json:"base_dn"`
	UserFilter         string `json:"user_filter"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	ActiveDirectory    bool   `json:"active_directory"`
	TestUsername       string `json:"test_username"`
}

func (r LdapTestRequest) toConfig() config.LdapConfig {
	return config.LdapConfig{
		Enabled:            true,
		Host:               r.Host,
		Port:               r.Port,
		UseTLS:             r.UseTLS,
		StartTLS:           r.StartTLS,
		BindDN:             r.BindDN,
		BindPassword:       r.BindPassword,
		BaseDN:             r.BaseDN,
		UserFilter:         r.UserFilter,
		InsecureSkipVerify: r.InsecureSkipVerify,
		ActiveDirectory:    r.ActiveDirectory,
	}
}

// LdapTestResult describes every step of the test so the panel can show what
// failed instead of a single generic error.
type LdapTestResult struct {
	Host                    string   `json:"host"`
	Port                    int      `json:"port"`
	Connected               bool     `json:"connected"`
	TLS                     bool     `json:"tls"`
	BindOK                  bool     `json:"bindOK"`
	ActiveDirectoryDetected bool     `json:"activeDirectoryDetected"`
	UsersFound              int      `json:"usersFound"`
	FirstDN                 string   `json:"firstDN"`
	FirstCN                 string   `json:"firstCN"`
	LatencyMs               int64    `json:"latencyMs"`
	Steps                   []string `json:"steps"`
	Error                   string   `json:"error"`
}

// TestLdap verifies the connection, the bind and the user search with the
// submitted values without saving anything.
func TestLdap(req LdapTestRequest) LdapTestResult {
	result := LdapTestResult{Host: req.Host, Port: req.Port}

	if strings.TrimSpace(req.Host) == "" {
		result.Error = "LDAP host is empty"
		result.Steps = append(result.Steps, result.Error)
		return result
	}

	if req.Port == 0 {
		result.Port = 389
	}

	cfg := req.toConfig()
	result.Steps = append(result.Steps, fmt.Sprintf("Connecting to %s:%d", req.Host, result.Port))

	start := time.Now()
	conn, err := connectLdap(cfg)
	if err != nil {
		result.Error = err.Error()
		result.Steps = append(result.Steps, result.Error)
		return result
	}
	result.Connected = true
	result.TLS = req.UseTLS || req.StartTLS
	result.Steps = append(result.Steps, fmt.Sprintf("Connected in %d ms%s", time.Since(start).Milliseconds(), tlsSuffix(result.TLS)))
	defer conn.Close()

	if err := bindServiceAccount(conn, cfg); err != nil {
		result.Error = err.Error()
		result.Steps = append(result.Steps, result.Error)
		return result
	}
	result.BindOK = true
	result.Steps = append(result.Steps, "Bind with the service account succeeded")

	if detected, err := detectActiveDirectory(conn); err == nil {
		result.ActiveDirectoryDetected = detected
		result.Steps = append(result.Steps, "Root DSE: Active Directory "+boolWord(detected))
	} else {
		result.Steps = append(result.Steps, "Could not read the root DSE: "+err.Error())
	}

	if strings.TrimSpace(req.BaseDN) == "" {
		result.Steps = append(result.Steps, "No base DN configured, the user search was skipped")
		return result
	}

	filter := "(|(objectClass=user)(objectClass=person))"
	if strings.TrimSpace(req.TestUsername) != "" {
		filter = strings.Replace(req.UserFilter, "%s", req.TestUsername, 1)
		result.Steps = append(result.Steps, fmt.Sprintf("Searching %q with the configured filter", req.TestUsername))
	} else {
		result.Steps = append(result.Steps, "No test username given, searching for any user object")
	}

	searchRequest := ldap.NewSearchRequest(
		req.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 5, 0, false,
		filter,
		[]string{"dn", "cn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		result.Error = err.Error()
		result.Steps = append(result.Steps, result.Error)
		return result
	}

	result.UsersFound = len(sr.Entries)
	if len(sr.Entries) > 0 {
		result.FirstDN = sr.Entries[0].DN
		result.FirstCN = sr.Entries[0].GetAttributeValue("cn")
		result.Steps = append(result.Steps, fmt.Sprintf("Found %d entries (first: %s)", len(sr.Entries), result.FirstDN))
	} else {
		result.Steps = append(result.Steps, "The search returned no entries")
	}

	return result
}

// detectActiveDirectory reads the root DSE, Microsoft AD answers with an
// organization objectClass while OpenLDAP answers with rootDSE.
func detectActiveDirectory(conn *ldap.Conn) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		"",
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		[]string{"objectClass", "dn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return false, err
	}

	for _, entry := range sr.Entries {
		for _, attribute := range entry.Attributes {
			if attribute.Name == "objectClass" {
				for _, value := range attribute.Values {
					if strings.EqualFold(value, "organization") {
						return true, nil
					}
				}
			}
		}
		if strings.Contains(strings.ToUpper(entry.DN), "DC=") {
			return true, nil
		}
	}

	return false, nil
}

func tlsSuffix(isTLS bool) string {
	if isTLS {
		return " over TLS"
	}
	return ""
}

func boolWord(value bool) string {
	if value {
		return "detected"
	}
	return "not detected"
}
