import { request } from '../request';

/**
 * Fetch the settings merged from server.yaml and the values saved in the panel
 */
export function fetchSettings() {
  return request<Api.Settings.Data>({ url: '/settings' });
}

/**
 * Save the settings, they take precedence over server.yaml
 */
export function saveSettings(data: Api.Settings.RuntimeSettings) {
  return request<Api.Settings.RuntimeSettings>({
    url: '/settings',
    method: 'POST',
    data
  });
}

/**
 * Reset the settings to the server.yaml defaults
 */
export function resetSettings() {
  return request<Api.Settings.RuntimeSettings>({ url: '/settings/reset', method: 'POST' });
}

/**
 * Test an LDAP bind with the values of the form
 */
export function testLdapBind(data: Api.Settings.LdapTestRequest) {
  return request<Api.Settings.LdapTestResult>({
    url: '/settings/ldap-test',
    method: 'POST',
    data
  });
}
