package management

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auth0/go-auth0"
)

func TestClient_Create(t *testing.T) {
	expectedClient := &Client{
		Name:        auth0.Stringf("Test Client (%s)", time.Now().Format(time.StampMilli)),
		Description: auth0.String("This is just a test client."),
	}

	err := m.Client.Create(expectedClient)

	assert.NoError(t, err)
	assert.NotEmpty(t, expectedClient.GetClientID())

	defer cleanupClient(t, expectedClient.GetClientID())
}

func TestClient_Read(t *testing.T) {
	expectedClient := givenAClient(t)
	defer cleanupClient(t, expectedClient.GetClientID())

	actualClient, err := m.Client.Read(expectedClient.GetClientID())

	assert.NoError(t, err)
	assert.Equal(t, expectedClient.GetName(), actualClient.GetName())
}

func TestClient_Update(t *testing.T) {
	expectedClient := givenAClient(t)
	defer cleanupClient(t, expectedClient.GetClientID())

	expectedDescription := "This is more than just a test client."
	expectedClient.Description = &expectedDescription

	clientID := expectedClient.GetClientID()
	expectedClient.ClientID = nil                       // Read-Only: Additional properties not allowed.
	expectedClient.SigningKeys = nil                    // Read-Only: Additional properties not allowed.
	expectedClient.JWTConfiguration.SecretEncoded = nil // Read-Only: Additional properties not allowed.

	err := m.Client.Update(clientID, expectedClient)

	assert.NoError(t, err)
	assert.Equal(t, expectedDescription, *expectedClient.Description)
}

func TestClient_Delete(t *testing.T) {
	expectedClient := givenAClient(t)

	err := m.Client.Delete(expectedClient.GetClientID())

	assert.NoError(t, err)

	actualClient, err := m.Client.Read(expectedClient.GetClientID())

	assert.Empty(t, actualClient)
	assert.EqualError(t, err, "404 Not Found: The client does not exist")
}

func TestClient_List(t *testing.T) {
	expectedClient := givenAClient(t)
	defer cleanupClient(t, expectedClient.GetClientID())

	clientList, err := m.Client.List(IncludeFields("client_id"))

	assert.NoError(t, err)
	assert.Contains(t, clientList.Clients, &Client{ClientID: expectedClient.ClientID})
}

func TestClient_RotateSecret(t *testing.T) {
	expectedClient := givenAClient(t)
	defer cleanupClient(t, expectedClient.GetClientID())

	oldSecret := expectedClient.GetClientSecret()
	actualClient, err := m.Client.RotateSecret(expectedClient.GetClientID())

	assert.NoError(t, err)
	assert.NotEqual(t, oldSecret, actualClient.GetClientSecret())
}

func TestJWTConfiguration(t *testing.T) {
	t.Run("MarshalJSON", func(t *testing.T) {
		for clientJWTConfiguration, expected := range map[*ClientJWTConfiguration]string{
			{}:                                   `{}`,
			{LifetimeInSeconds: auth0.Int(1000)}: `{"lifetime_in_seconds":1000}`,
		} {
			jsonBody, err := json.Marshal(clientJWTConfiguration)
			assert.NoError(t, err)
			assert.Equal(t, string(jsonBody), expected)
		}
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		for jsonBody, expected := range map[string]*ClientJWTConfiguration{
			`{}`:                             {LifetimeInSeconds: nil},
			`{"lifetime_in_seconds":1000}`:   {LifetimeInSeconds: auth0.Int(1000)},
			`{"lifetime_in_seconds":"1000"}`: {LifetimeInSeconds: auth0.Int(1000)},
		} {
			var actual ClientJWTConfiguration
			err := json.Unmarshal([]byte(jsonBody), &actual)

			assert.NoError(t, err)
			assert.Equal(t, &actual, expected)
		}
	})
}

func givenAClient(t *testing.T) *Client {
	client := &Client{
		Name:        auth0.Stringf("Test Client (%s)", time.Now().Format(time.StampMilli)),
		Description: auth0.String("This is just a test client."),
	}

	err := m.Client.Create(client)
	require.NoError(t, err)

	return client
}

func cleanupClient(t *testing.T, clientID string) {
	err := m.Client.Delete(clientID)
	require.NoError(t, err)
}
