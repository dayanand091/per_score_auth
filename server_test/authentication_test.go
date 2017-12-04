package identity_server_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/wrsinc/gogenproto/identity/enums/partystatus"
	"github.com/wrsinc/gogenproto/identity/enums/publiccredentialtype"
	identity_service "github.com/wrsinc/gogenproto/identity/service"
	"github.com/wrsinc/gogenproto/identity/service/credentials"
	"github.com/wrsinc/gogenproto/identity/service/party"
	h "github.com/wrsinc/identity/testhelpers"
	"golang.org/x/net/context"
)

func TestServer_TestTwoTenantAuthenticationSameEmail(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {

		email0 := h.GenerateEmail()
		tenant0 := "WRS-" + uuid.New().String()
		createPartyRequest0 := h.CreatePartyRequestWithTenantAndEmail(tenant0, email0)
		createPartyResponse0 := CreateParty(t, client, createPartyRequest0)

		tenant1 := "EvilCorp-" + uuid.New().String()
		createPartyRequest1 := h.CreatePartyRequestWithTenantAndEmail(tenant1, email0)
		createPartyResponse1 := CreateParty(t, client, createPartyRequest1)

		if createPartyResponse0.PartyId == createPartyResponse1.PartyId {
			t.Fatalf("Should be 2 different parties.")
		}

		{
			createCredentialsRequest := h.CreatePublicCredentialRequest(publiccredentialtype.PublicCredentialType_EMAIL,
				email0, "password0", createPartyResponse0.PartyId, "")
			_, err := client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
			if err != nil {
				t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
			}
		}

		{
			createCredentialsRequest := h.CreatePublicCredentialRequest(publiccredentialtype.PublicCredentialType_EMAIL,
				email0, "password1", createPartyResponse1.PartyId, "")
			_, err := client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
			if err != nil {
				t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
			}
		}

		authenticateRequest := &credentials.AuthenticateRequest{
			PublicCredentialType: publiccredentialtype.PublicCredentialType_EMAIL,
			Value:                email0,
		}

		{
			authenticateRequest.TenantId = ""
			authenticateRequest.Password = "password0"
			authenticateResponse, err := client.Authenticate(context.Background(), authenticateRequest)
			switch {
			case err != nil:
				t.Fatalf("Failed to call Authenticate %+v", err)
			case authenticateResponse.Status != credentials.AuthenticateResponse_INVALID_REQUEST:
				t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_INVALID_REQUEST, authenticateResponse.Status)
			}
		}

		{
			authenticateRequest.TenantId = tenant0
			authenticateRequest.Password = "password1"
			authenticateResponse, err := client.Authenticate(context.Background(), authenticateRequest)
			switch {
			case err != nil:
				t.Fatalf("Failed to call Authenticate %+v", err)
			case authenticateResponse.Status != credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS:
				t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS, authenticateResponse.Status)
			}

			authenticateRequest.Password = "password0"
			authenticateResponse, err = client.Authenticate(context.Background(), authenticateRequest)
			switch {
			case err != nil:
				t.Fatalf("Failed to call Authenticate %+v", err)
			case authenticateResponse.Status != credentials.AuthenticateResponse_SUCCESS:
				t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_SUCCESS, authenticateResponse.Status)
			case authenticateResponse.PartyId != createPartyResponse0.PartyId:
				t.Fatalf("Invalid PartyId, expected %s, got %s", createPartyResponse0.PartyId, authenticateResponse.PartyId)
			}
		}
		{
			authenticateRequest.TenantId = tenant1
			authenticateRequest.Password = "password0"
			authenticateResponse, err := client.Authenticate(context.Background(), authenticateRequest)
			switch {
			case err != nil:
				t.Fatalf("Failed to call Authenticate %+v", err)
			case authenticateResponse.Status != credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS:
				t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS, authenticateResponse.Status)
			}

			authenticateRequest.Password = "password1"
			authenticateResponse, err = client.Authenticate(context.Background(), authenticateRequest)
			switch {
			case err != nil:
				t.Fatalf("Failed to call Authenticate %+v", err)
			case authenticateResponse.Status != credentials.AuthenticateResponse_SUCCESS:
				t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_SUCCESS, authenticateResponse.Status)
			case authenticateResponse.PartyId != createPartyResponse1.PartyId:
				t.Fatalf("Invalid PartyId, expected %s, got %s", createPartyResponse1.PartyId, authenticateResponse.PartyId)
			}
		}

		return nil
	}, t)
}

func TestServer_Authenticate(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		authenticateRequest := &credentials.AuthenticateRequest{
			PublicCredentialType: publiccredentialtype.PublicCredentialType_EMAIL,
		}
		authenticateResponse, err := client.Authenticate(context.Background(), authenticateRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call Authenticate %+v", err)
		case authenticateResponse.Status != credentials.AuthenticateResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_INVALID_REQUEST, authenticateResponse.Status)
		}

		// create Party
		req := h.CreatePartyRequest()

		createPartyResponse := CreateParty(t, client, req)
		t.Logf("createPartyResponse: %+v", createPartyResponse)

		createCredentialsRequest := h.CreatePublicCredentialRequest(publiccredentialtype.PublicCredentialType_EMAIL,
			"", "", createPartyResponse.PartyId, "")
		createCredentialsResponse, err := client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
		}
		t.Logf("createCredentialsResponse: %+v", createCredentialsResponse)

		authenticateRequest.Value = createCredentialsRequest.Value
		authenticateRequest.TenantId = req.TenantId
		authenticateResponse, err = client.Authenticate(context.Background(), authenticateRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call Authenticate %+v", err)
		case authenticateResponse.Status != credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS, authenticateResponse.Status)
		}

		authenticateRequest.Password = createCredentialsRequest.Password
		authenticateRequest.TenantId = req.TenantId
		authenticateResponse, err = client.Authenticate(context.Background(), authenticateRequest)
		t.Logf("authenticateRequest %+v", authenticateRequest)
		t.Logf("authenticateResponse %+v", authenticateResponse)
		switch {
		case err != nil:
			t.Fatalf("Failed to call Authenticate %+v", err)
		case authenticateResponse.Status != credentials.AuthenticateResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_SUCCESS, authenticateResponse.Status)
		case authenticateResponse.PartyId != createPartyResponse.PartyId:
			t.Fatalf("Invalid PartyId, expected %s, got %s", createPartyResponse.PartyId, authenticateResponse.PartyId)
		}

		authenticateRequest.TenantId = req.TenantId + "_does_not_exist"
		authenticateResponse, err = client.Authenticate(context.Background(), authenticateRequest)
		t.Logf("authenticateRequest %+v", authenticateRequest)
		t.Logf("authenticateResponse %+v", authenticateResponse)
		switch {
		case err != nil:
			t.Fatalf("Failed to call Authenticate %+v", err)
		case authenticateResponse.Status != credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS, authenticateResponse.Status)
		case authenticateResponse.PartyId != "":
			t.Fatalf("Invalid PartyId, expected %s, got %s", "", authenticateResponse.PartyId)
		}

		in := &party.UpdatePartyStatusRequest{
			PartyId:     createPartyResponse.PartyId,
			PartyStatus: partystatus.PartyStatus_INACTIVE,
		}

		updatePartyResponse, err := client.UpdatePartyStatus(context.Background(), in)
		if err != nil {
			t.Fatalf("Failed to call UpdatePartyStatus: %+v", err)
		}
		t.Logf("updatePartyResponse: %+v", updatePartyResponse)

		authenticateRequest.TenantId = req.TenantId
		authenticateResponse, err = client.Authenticate(context.Background(), authenticateRequest)
		t.Logf("authenticateRequest %+v", authenticateRequest)
		t.Logf("authenticateResponse %+v", authenticateResponse)
		switch {
		case err != nil:
			t.Fatalf("Failed to call Authenticate %+v", err)
		case authenticateResponse.Status == credentials.AuthenticateResponse_SUCCESS:
			t.Fatalf("Invalid status, expected not %s, got %s", credentials.AuthenticateResponse_SUCCESS, authenticateResponse.Status)
		case authenticateResponse.PartyId != createPartyResponse.PartyId:
			t.Fatalf("Invalid PartyId, expected %s, got %s", createPartyResponse.PartyId, authenticateResponse.PartyId)
		}

		return nil
	}, t)
}

func TestServer_AuthenticateAndRetrieve(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		authenticateRequest := &credentials.AuthenticateRequest{
			PublicCredentialType: publiccredentialtype.PublicCredentialType_EMAIL,
		}
		authenticateResponse, err := client.AuthenticateAndRetrieve(context.Background(), authenticateRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		case authenticateResponse.Status != credentials.RetrievePartyResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.RetrievePartyResponse_INVALID_REQUEST, authenticateResponse.Status)
		}

		req := h.CreatePartyRequest()

		// create Party
		createPartyResponse := CreateParty(t, client, req)
		t.Logf("createPartyResponse: %+v", createPartyResponse)

		createCredentialsRequest := h.CreatePublicCredentialRequest(publiccredentialtype.PublicCredentialType_EMAIL,
			"", "", createPartyResponse.PartyId, "")

		createCredentialsResponse, err := client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
		}
		t.Logf("createCredentialsResponse: %+v", createCredentialsResponse)

		authenticateRequest.TenantId = req.TenantId
		authenticateRequest.Value = createCredentialsRequest.Value
		authenticateResponse, err = client.AuthenticateAndRetrieve(context.Background(), authenticateRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		case authenticateResponse.Status != credentials.RetrievePartyResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.AuthenticateResponse_CREDENTIAL_NOT_EXISTS, authenticateResponse.Status)
		}

		authenticateRequest.Password = createCredentialsRequest.Password
		authenticateResponse, err = client.AuthenticateAndRetrieve(context.Background(), authenticateRequest)
		t.Logf("authenticateRequest %+v", authenticateRequest)
		t.Logf("authenticateResponse %+v", authenticateResponse)
		switch {
		case err != nil:
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		case authenticateResponse.Status != credentials.RetrievePartyResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.RetrievePartyResponse_SUCCESS, authenticateResponse.Status)
		case authenticateResponse.Party == nil:
			t.Fatalf("Party, expected non-empty, got %+v", authenticateResponse.Party)
		}

		authenticateRequest.Password = createCredentialsRequest.Password
		authenticateRequest.TenantId = req.TenantId + "does_not_exist"
		authenticateResponse, err = client.AuthenticateAndRetrieve(context.Background(), authenticateRequest)
		t.Logf("authenticateRequest %+v", authenticateRequest)
		t.Logf("authenticateResponse %+v", authenticateResponse)
		switch {
		case err != nil:
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		case authenticateResponse.Status != credentials.RetrievePartyResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials.RetrievePartyResponse_CREDENTIAL_NOT_EXISTS, authenticateResponse.Status)
		case authenticateResponse.Party != nil:
			t.Fatalf("Party, expected nil got %+v", authenticateResponse.Party)
		}

		return nil
	}, t)
}
