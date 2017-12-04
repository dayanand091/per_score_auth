package identity_server_test

import (
	"testing"

	"github.com/google/uuid"
	databunker_service "github.com/wrsinc/gogenproto/databunker/service"
	databunker_service_type "github.com/wrsinc/gogenproto/databunker/service/databunker"
	"github.com/wrsinc/gogenproto/identity/enums/partyownertype"
	"github.com/wrsinc/gogenproto/identity/enums/partystatus"
	"github.com/wrsinc/gogenproto/identity/enums/privatecredentialtype"
	"github.com/wrsinc/gogenproto/identity/enums/publiccredentialtype"
	"github.com/wrsinc/gogenproto/identity/enums/roletype"
	identity_service "github.com/wrsinc/gogenproto/identity/service"
	common_service_type "github.com/wrsinc/gogenproto/identity/service/common"
	credentials_service_type "github.com/wrsinc/gogenproto/identity/service/credentials"
	party_service_type "github.com/wrsinc/gogenproto/identity/service/party"
	party_relationship_service_type "github.com/wrsinc/gogenproto/identity/service/partyrelationship"
	role_service_type "github.com/wrsinc/gogenproto/identity/service/role"
	"github.com/wrsinc/identity/params"
	h "github.com/wrsinc/identity/testhelpers"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestServer_CreatePartyWithDupePrivateCredentialsSameAndDifferentTenant(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		createPartyWithCredentialsRequest, createPartyWithCredentialsResponse, err := CreatePartyWithCredentials(t, client)
		validateCreatePartyWithCredentials(t, createPartyWithCredentialsRequest, createPartyWithCredentialsResponse, err)

		// create another set of the same private credentials in the same tenant (should fail)
		{
			req := h.CreatePartyRequestWithTenantAndEmail(createPartyWithCredentialsRequest.GetCreatePartyRequest().TenantId, h.GenerateEmail())

			createPartyResponse1 := CreateParty(t, client, req)
			createPrivateCredentialsRequest := &credentials_service_type.CreatePrivateCredentialsRequest{
				PartyId:               createPartyResponse1.PartyId,
				RoleId:                createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].RoleId,
				PrivateCredentialType: createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].PrivateCredentialType,
				Value: &common_service_type.PrivateCredentialValue{
					Valuetype: &common_service_type.PrivateCredentialValue_StringValue{
						StringValue: createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].Value.GetStringValue(),
					},
				},
			}

			createPrivateCredentialsResponse, err := client.CreatePrivateCredentials(context.Background(), createPrivateCredentialsRequest)
			t.Logf("req: %+v", createPrivateCredentialsRequest)
			t.Logf("resp: %+v", createPrivateCredentialsResponse)
			switch {
			case err != nil:
				t.Fatalf("Failed to call CreatePrivateCredentials: %+v", err)
			case createPrivateCredentialsResponse.Status != credentials_service_type.CreateCredentialsResponse_DUPLICATE_CREDENTIAL:
				t.Fatalf("Invalid Status, expected %s, got, %s", credentials_service_type.CreateCredentialsResponse_DUPLICATE_CREDENTIAL, createPrivateCredentialsResponse.Status)
			}
		}

		// create another set of the same private credentials in a different tenant (should succeed)
		{
			req := h.CreatePartyRequestWithTenantAndEmail("EvilCorp0-"+uuid.New().String(), h.GenerateEmail())

			createPartyResponse1 := CreateParty(t, client, req)
			createPrivateCredentialsRequest := &credentials_service_type.CreatePrivateCredentialsRequest{
				PartyId:               createPartyResponse1.PartyId,
				RoleId:                createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].RoleId,
				PrivateCredentialType: createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].PrivateCredentialType,
				Value: &common_service_type.PrivateCredentialValue{
					Valuetype: &common_service_type.PrivateCredentialValue_StringValue{
						StringValue: createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].Value.GetStringValue(),
					},
				},
			}

			createPrivateCredentialsResponse, err := client.CreatePrivateCredentials(context.Background(), createPrivateCredentialsRequest)
			t.Logf("req: %+v", createPrivateCredentialsRequest)
			t.Logf("resp: %+v", createPrivateCredentialsResponse)
			switch {
			case err != nil:
				t.Fatalf("Failed to call CreatePrivateCredentials: %+v", err)
			case createPrivateCredentialsResponse.Status != credentials_service_type.CreateCredentialsResponse_SUCCESS:
				t.Fatalf("Invalid Status, expected %s, got, %s", credentials_service_type.CreateCredentialsResponse_SUCCESS, createPrivateCredentialsResponse.Status)
			}
		}

		return nil
	}, t)
}

func TestServer_RetrieveParty(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		retrievePartyRequest := &party_service_type.RetrievePartyRequest{
			PartyId: uuid.New().String(),
		}

		// retrieve non-existent Party
		retrievePartyResponse, err := client.RetrieveParty(context.Background(), retrievePartyRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		case retrievePartyResponse.Status != party_service_type.RetrievePartyResponse_PARTY_NOT_EXISTS:
			t.Fatalf("Invalid Status, expected %s, got, %s",
				party_service_type.RetrievePartyResponse_PARTY_NOT_EXISTS, retrievePartyResponse.Status)
		case retrievePartyResponse.Party != nil:
			t.Fatalf("Invalid Party, expected nil, got, %v", retrievePartyResponse.Party)
		}

		createPartyRequest := h.CreatePartyRequest()
		createPartyResponse := CreateParty(t, client, createPartyRequest)

		// retrieve only the Party
		retrievePartyRequest.PartyId = createPartyResponse.PartyId
		retrievePartyResponse, err = client.RetrieveParty(context.Background(), retrievePartyRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		case retrievePartyResponse.Status != party_service_type.RetrievePartyResponse_SUCCESS:
			t.Fatalf("Invalid Status, expected %s, got, %s",
				party_service_type.RetrievePartyResponse_SUCCESS, retrievePartyResponse.Status)
		case retrievePartyResponse.Party == nil:
			t.Fatalf("Invalid Party, expected not nil, got, %v", retrievePartyResponse.Party)
		case retrievePartyResponse.Party.PartyOwnerType == partyownertype.PartyOwnerType_UNKNOWN:
			t.Fatalf("Invalid PartyOwnerType got, %s", retrievePartyResponse.Party.PartyOwnerType)
		case retrievePartyResponse.Party.PartyOwnerType != partyownertype.PartyOwnerType_UNKNOWN && retrievePartyResponse.Party.PartyOwner == "":
			t.Fatalf("Invalid PartyOwner got, %s", retrievePartyResponse.Party.PartyOwner)
		}

		type roleElement struct {
			request  *role_service_type.CreateRoleRequest
			response *role_service_type.CreateRoleResponse
		}
		roles := make(map[string]roleElement)
		createRoleRequest1 := h.CreateRoleRequest(createPartyResponse.PartyId, roletype.RoleType_EMPLOYEE)
		createRoleResponse1 := CreateRole(t, client, createRoleRequest1)
		roles[createRoleResponse1.RoleId] = roleElement{createRoleRequest1, createRoleResponse1}

		// Retrieve Party and single Role
		retrievePartyResponse, err = client.RetrieveParty(context.Background(), retrievePartyRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		case retrievePartyResponse.Status != party_service_type.RetrievePartyResponse_SUCCESS:
			t.Fatalf("Invalid Status, expected %s, got, %s",
				party_service_type.RetrievePartyResponse_SUCCESS, retrievePartyResponse.Status)
		case len(retrievePartyResponse.Roles) != 1:
			t.Fatalf("Invalid Roles, expected 1, got, %d", len(retrievePartyResponse.Roles))
		case retrievePartyResponse.Roles[0].RoleId != createRoleResponse1.RoleId:
			t.Fatalf("Invalid roleId, expected %s, got %s", createRoleResponse1.RoleId, retrievePartyResponse.Roles[0].RoleId)
		}

		createRoleRequest2 := h.CreateRoleRequest(createPartyResponse.PartyId, roletype.RoleType_AUDIENCE_PROVIDER)
		createRoleResponse2 := CreateRole(t, client, createRoleRequest2)
		roles[createRoleResponse2.RoleId] = roleElement{createRoleRequest2, createRoleResponse2}

		// Retrieve Party and two Roles
		retrievePartyResponse, err = client.RetrieveParty(context.Background(), retrievePartyRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		case retrievePartyResponse.Status != party_service_type.RetrievePartyResponse_SUCCESS:
			t.Fatalf("Invalid Status, expected %s, got, %s",
				party_service_type.RetrievePartyResponse_SUCCESS, retrievePartyResponse.Status)
		case len(retrievePartyResponse.Roles) != 2:
			t.Fatalf("Invalid Roles, expected 2, got, %d", len(retrievePartyResponse.Roles))
		}
		for _, r := range retrievePartyResponse.Roles {
			if _, present := roles[r.RoleId]; !present {
				t.Fatal("Missing role")
			}
		}

		// create Public Credential
		createPublicCredentialsRequest := h.CreatePublicCredentialRequest(
			publiccredentialtype.PublicCredentialType_USERNAME_PASSWORD, "",
			"", createPartyResponse.PartyId, createRoleResponse2.RoleId)
		createPublicCredentialsResponse, err := client.CreatePublicCredentials(context.Background(), createPublicCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
		}

		// create Private Credential
		createPrivateCredentialsRequest, _ := h.CreatePrivateCredentialRequest(
			privatecredentialtype.PrivateCredentialType_BIOMETRIC_TOKEN, "", createPartyResponse.PartyId)
		createPrivateCredentialsResponse, err := client.CreatePrivateCredentials(context.Background(), createPrivateCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		}

		// Create another Party and PartyRelationship
		createOtherPartyRequest := h.CreatePartyRequest()
		createOtherPartyResponse := CreateParty(t, client, createOtherPartyRequest)

		createPartyRelationship := CreatePartyRelationship(createPartyResponse.PartyId, createOtherPartyResponse.PartyId)
		createPartyRelationshipResponse, err := client.CreatePartyRelationship(context.Background(), createPartyRelationship)
		if err != nil {
			t.Fatalf("Failed to call CreateParty: %+v", err)
		}

		retrievePartyResponse, err = client.RetrieveParty(context.Background(), retrievePartyRequest)
		t.Logf("retrievePartyResponse: %+v", retrievePartyResponse)

		// TODO: move this and the associated test into a test for CreatePublicCredentials
		conn, err := grpc.Dial(params.DataBunkerAddress, grpc.WithInsecure())
		if err != nil {
			t.Fatalf("did not connect: %+v", err)
		}
		defer conn.Close()
		databunkerClient := databunker_service.NewDatabunkerClient(conn)

		t.Logf("retrievePartyResponse: %+v", retrievePartyResponse)
		publicResponse, err := databunkerClient.Retrieve(context.Background(), &databunker_service_type.RetrieveRequest{
			Id: retrievePartyResponse.PublicCredentials[0].PasswordId,
		})
		if err != nil {
			t.Fatalf("Failed to call Data Bunker: %+v", err)
		}

		privateResponse, err := databunkerClient.Retrieve(context.Background(), &databunker_service_type.RetrieveRequest{
			Id: retrievePartyResponse.PrivateCredentials[0].PasswordId,
		})
		if err != nil {
			t.Fatalf("Failed to call Data Bunker: %+v", err)
		}

		switch {
		case err != nil:
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		case retrievePartyResponse.Status != party_service_type.RetrievePartyResponse_SUCCESS:
			t.Fatalf("Invalid Status, expected %s, got, %s",
				party_service_type.RetrievePartyResponse_SUCCESS, retrievePartyResponse.Status)

			// Public Credentials
		case len(retrievePartyResponse.PublicCredentials) != 1:
			t.Fatalf("Invalid Public Credential count, expected 1, got, %d", len(retrievePartyResponse.PublicCredentials))
		case retrievePartyResponse.PublicCredentials[0].CredentialId != createPublicCredentialsResponse.CredentialsId:
			t.Fatalf("Invalid Public Credential Password, expected %s, got, %s",
				createPublicCredentialsResponse.CredentialsId, retrievePartyResponse.PublicCredentials[0].CredentialId)
		case string(publicResponse.Data) != createPublicCredentialsRequest.Password:
			// TODO: move this into a test for CreatePublicCredentials
			t.Fatalf("Invalid Public Credential Password, expected %s, got, %s",
				createPublicCredentialsRequest.Password, string(publicResponse.Data))

			// Private Credentials
		case len(retrievePartyResponse.PrivateCredentials) != 1:
			t.Fatalf("Invalid Private Credential count, expected 1, got, %d", len(retrievePartyResponse.PrivateCredentials))
		case retrievePartyResponse.PrivateCredentials[0].CredentialId != createPrivateCredentialsResponse.CredentialsId:
			t.Fatalf("Invalid Private Credential Id, expected %s, got, %s",
				createPrivateCredentialsResponse.CredentialsId, retrievePartyResponse.PrivateCredentials[0].CredentialId)
		case string(privateResponse.Data) != createPrivateCredentialsRequest.Password:
			t.Fatalf("Invalid Private Credential Password, expected %s, got, %s",
				createPrivateCredentialsRequest.Password, string(privateResponse.Data))

			// PartyRelationship
		case len(retrievePartyResponse.PartyRelationships) != 1:
			t.Fatalf("Invalid PartyRelationships count, expected 1, got, %d", len(retrievePartyResponse.PartyRelationships))
		case retrievePartyResponse.PartyRelationships[0].PartyRelationshipId != createPartyRelationshipResponse.PartyRelationshipId:
			t.Fatalf("Invalid PartyRelationships Password, expected %s, got, %s",
				createPartyRelationshipResponse.PartyRelationshipId, retrievePartyResponse.PartyRelationships[0].PartyRelationshipId)
		}

		return nil
	}, t)
}

func TestServer_CreateAnonymousParty(t *testing.T) {
	// TODO(tkuchlein): Need to implement
}

func TestServer_CreateRoleAndPrivateCredential(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		createPartyResponse := CreateParty(t, client, h.CreatePartyRequest())
		t.Logf("createPartyResponse: %+v", createPartyResponse)

		roleRequest := h.CreateRoleRequest(createPartyResponse.PartyId, roletype.RoleType_FACEBOOK_MESSENGER_USER)
		privateCredentialsRequest, _ := h.CreatePrivateCredentialRequest(
			privatecredentialtype.PrivateCredentialType_BIOMETRIC_TOKEN, "", "")
		createRequest := &role_service_type.CreateRoleAndPrivateCredentialRequest{
			CreateRoleRequest:               roleRequest,
			CreatePrivateCredentialsRequest: privateCredentialsRequest,
		}
		createResponse, err := client.CreateRoleAndPrivateCredential(context.Background(), createRequest)
		t.Logf("response: %+v", createResponse)
		switch {
		case err != nil:
			t.Fatalf("Error response %+v", err)
		case createResponse.Status != role_service_type.CreateRoleAndPrivateCredentialResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s got %s", role_service_type.CreateRoleAndPrivateCredentialResponse_SUCCESS, createResponse.Status)
		}

		createResponse, err = client.CreateRoleAndPrivateCredential(context.Background(), createRequest)
		t.Logf("response: %+v", createResponse)

		return nil
	}, t)
}

func TestServer_CreatePartyRelationship(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		// test that having no parties correctly fails
		createPartyRelationship := CreatePartyRelationship(uuid.New().String(), uuid.New().String())
		createPartyRelationshipResponse, err := client.CreatePartyRelationship(context.Background(), createPartyRelationship)
		if err != nil {
			t.Fatalf("Failed to call CreateParty: %+v", err)
		}
		validatePartyRelationship(t, createPartyRelationshipResponse, party_relationship_service_type.CreatePartyRelationshipResponse_PARTY_FROM_NOT_EXISTS, err)

		// create PartyFrom
		createPartyFromResponse := CreateParty(t, client, h.CreatePartyRequest())
		t.Logf("createPartyResponse: %+v", createPartyFromResponse)

		// Validate that it fails on missing PartyTo
		createPartyRelationship.PartyIdFrom = createPartyFromResponse.PartyId
		createPartyRelationshipResponse, err = client.CreatePartyRelationship(context.Background(), createPartyRelationship)
		if err != nil {
			t.Fatalf("Failed to call CreateParty: %+v", err)
		}
		validatePartyRelationship(t, createPartyRelationshipResponse, party_relationship_service_type.CreatePartyRelationshipResponse_PARTY_TO_NOT_EXISTS, err)

		// create PartyTo
		createPartyToResponse := CreateParty(t, client, h.CreatePartyRequest())
		t.Logf("createPartyResponse: %+v", createPartyToResponse)

		// Validate that it succeeds with a valid request
		createPartyRelationship.PartyIdTo = createPartyToResponse.PartyId
		createPartyRelationshipResponse, err = client.CreatePartyRelationship(context.Background(), createPartyRelationship)
		if err != nil {
			t.Fatalf("Failed to call CreatePartyRelationship: %+v", err)
		}
		validatePartyRelationship(t, createPartyRelationshipResponse, party_relationship_service_type.CreatePartyRelationshipResponse_SUCCESS, err)

		// Validate that it fails with a second attempt to create the same relationship
		createPartyRelationshipResponse, err = client.CreatePartyRelationship(context.Background(), createPartyRelationship)
		if err != nil {
			t.Fatalf("Failed to call CreatePartyRelationship: %+v", err)
		}
		validatePartyRelationship(t, createPartyRelationshipResponse, party_relationship_service_type.CreatePartyRelationshipResponse_PARTY_RELATIONSHIP_EXISTS, err)

		return nil
	}, t)
}

func TestServer_CreatePublicCredentials(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		createCredentialsRequest := h.CreatePublicCredentialRequest(
			publiccredentialtype.PublicCredentialType_USERNAME_PASSWORD, "", "", "", "non-existent-role")

		createCredentialsResponse, err := client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
		}
		validateCreateCredentials(t, createCredentialsResponse, credentials_service_type.CreateCredentialsResponse_PARTY_NOT_EXISTS, err)

		// create PartyFrom
		createPartyResponse := CreateParty(t, client, h.CreatePartyRequest())
		t.Logf("createPartyResponse: %+v", createPartyResponse)

		createCredentialsRequest.PartyId = createPartyResponse.PartyId
		createCredentialsResponse, err = client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
		}
		validateCreateCredentials(t, createCredentialsResponse, credentials_service_type.CreateCredentialsResponse_ROLE_NOT_EXISTS, err)

		createCredentialsRequest.RoleId = ""
		createCredentialsResponse, err = client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePublicCredentials %+v", err)
		}
		validateCreateCredentials(t, createCredentialsResponse, credentials_service_type.CreateCredentialsResponse_SUCCESS, err)

		password := createCredentialsRequest.Password
		createCredentialsRequest.Password = ""
		createCredentialsResponse, err = client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err == nil {
			t.Fatal("Expected failure but got success")
		}

		createCredentialsRequest.Password = password
		createCredentialsRequest.Value = ""
		createCredentialsResponse, err = client.CreatePublicCredentials(context.Background(), createCredentialsRequest)
		if err == nil {
			t.Fatal("Expected failure but got success")
		}

		return nil
	}, t)
}

func TestServer_CreatePrivateCredentials(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		createCredentialsRequest, _ := h.CreatePrivateCredentialRequest(
			privatecredentialtype.PrivateCredentialType_BIOMETRIC_TOKEN, "", "")
		createCredentialsResponse, err := client.CreatePrivateCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		}
		validateCreateCredentials(t, createCredentialsResponse, credentials_service_type.CreateCredentialsResponse_PARTY_NOT_EXISTS, err)

		// create Party
		createPartyResponse := CreateParty(t, client, h.CreatePartyRequest())
		t.Logf("createPartyResponse: %+v", createPartyResponse)

		createCredentialsRequest.PartyId = createPartyResponse.PartyId
		t.Logf("createCredentialsRequest: %+v", createCredentialsRequest)
		createCredentialsResponse, err = client.CreatePrivateCredentials(context.Background(), createCredentialsRequest)
		if err != nil {
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		}
		validateCreateCredentials(t, createCredentialsResponse, credentials_service_type.CreateCredentialsResponse_SUCCESS, err)

		password := createCredentialsRequest.Password
		createCredentialsRequest.Password = ""
		createCredentialsResponse, err = client.CreatePrivateCredentials(context.Background(), createCredentialsRequest)
		if err == nil {
			t.Fatal("Expected failure but got success")
		}

		createCredentialsRequest.Password = password
		createCredentialsRequest.Value = &common_service_type.PrivateCredentialValue{
			Valuetype: &common_service_type.PrivateCredentialValue_StringValue{
				StringValue: "",
			},
		}

		createCredentialsResponse, err = client.CreatePrivateCredentials(context.Background(), createCredentialsRequest)
		if err == nil {
			t.Fatal("Expected failure but got success")
		}

		return nil
	}, t)
}

func TestServer_DeletePublicCredentialsByTypeValue(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		// invalid request
		deletePublicCredentialsRequest := &credentials_service_type.DeletePublicCredentialsRequest{
			Type: &credentials_service_type.DeletePublicCredentialsRequest_TypeValue{
				TypeValue: &credentials_service_type.DeletePublicCredentialsRequest_ByTypeValue{
					PublicCredentialType: publiccredentialtype.PublicCredentialType_USERNAME_PASSWORD,
					Value:                "",
				},
			},
		}
		deleteCredentialsResponse, err := client.DeletePublicCredentials(context.Background(), deletePublicCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST, deleteCredentialsResponse.Status)
		}

		// create Party and credentials
		createPartyWithCredentialsRequest, createPartyWithCredentialsResponse, _ := CreatePartyWithCredentials(t, client)

		// Delete credential
		deletePublicCredentialsRequest.Type =
			&credentials_service_type.DeletePublicCredentialsRequest_TypeValue{
				TypeValue: &credentials_service_type.DeletePublicCredentialsRequest_ByTypeValue{
					Value: createPartyWithCredentialsRequest.CreatePublicCredentialsRequests[0].Value,
				},
			}
		deleteCredentialsResponse, err = client.DeletePublicCredentials(context.Background(), deletePublicCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST, deleteCredentialsResponse.Status)
		}

		// valid request
		deletePublicCredentialsRequest.Type =
			&credentials_service_type.DeletePublicCredentialsRequest_TypeValue{
				TypeValue: &credentials_service_type.DeletePublicCredentialsRequest_ByTypeValue{
					Value:                createPartyWithCredentialsRequest.CreatePublicCredentialsRequests[0].Value,
					PublicCredentialType: publiccredentialtype.PublicCredentialType_USERNAME_PASSWORD,
					PartyId:              createPartyWithCredentialsResponse.PartyId + "does_not_exist",
				},
			}

		// invalid party to delete
		deleteCredentialsResponse, err = client.DeletePublicCredentials(context.Background(), deletePublicCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS, deleteCredentialsResponse.Status)
		}

		// valid party (actually delete it)
		deletePublicCredentialsRequest.GetTypeValue().PartyId = createPartyWithCredentialsResponse.PartyId
		deleteCredentialsResponse, err = client.DeletePublicCredentials(context.Background(), deletePublicCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_SUCCESS, deleteCredentialsResponse.Status)
		}

		// and again to prove that it fails the second time
		deleteCredentialsResponse, err = client.DeletePublicCredentials(context.Background(), deletePublicCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS, deleteCredentialsResponse.Status)
		}

		return nil
	}, t)
}

func TestServer_DeletePublicCredentialsById(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		deletePublicCredentialsByIdRequest := &credentials_service_type.DeletePublicCredentialsRequest{
			Type: &credentials_service_type.DeletePublicCredentialsRequest_CredentialsId{
				CredentialsId: "",
			},
		}
		deleteCredentialsByIdResponse, err := client.DeletePublicCredentials(context.Background(), deletePublicCredentialsByIdRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsByIdResponse.Status != credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST, deleteCredentialsByIdResponse.Status)
		}

		// create Party and credentials
		_, createPartyWithCredentialsResponse, _ := CreatePartyWithCredentials(t, client)

		// Delete credential
		switch x := deletePublicCredentialsByIdRequest.Type.(type) {
		case *credentials_service_type.DeletePublicCredentialsRequest_CredentialsId:
			x.CredentialsId = createPartyWithCredentialsResponse.PublicCredentialIds[0]
		}
		//credentials_service_type.DeleteCredentialsRequest_CredentialsId(deletePublicCredentialsByIdRequest.Type).CredentialsId = createPartyWithCredentialsResponse.PublicCredentialIds[0]
		deleteCredentialsByIdResponse, err = client.DeletePublicCredentials(context.Background(), deletePublicCredentialsByIdRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsByIdResponse.Status != credentials_service_type.DeleteCredentialsResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_SUCCESS, deleteCredentialsByIdResponse.Status)
		}

		deleteCredentialsByIdResponse, err = client.DeletePublicCredentials(context.Background(), deletePublicCredentialsByIdRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePublicCredentials %+v", err)
		case deleteCredentialsByIdResponse.Status != credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS, deleteCredentialsByIdResponse.Status)
		}

		return nil
	}, t)
}

func TestServer_DeletePrivateCredentials(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		deletePrivateCredentialsRequest := &credentials_service_type.DeletePrivateCredentialsRequest{
			Type: &credentials_service_type.DeletePrivateCredentialsRequest_TypeValue{
				TypeValue: &credentials_service_type.DeletePrivateCredentialsRequest_ByTypeValue{
					PrivateCredentialType: privatecredentialtype.PrivateCredentialType_BIOMETRIC_TOKEN,
					Value: "",
				},
			},
		}
		deleteCredentialsResponse, err := client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePrivateCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST, deleteCredentialsResponse.Status)
		}

		// create Party and credentials
		createPartyWithCredentialsRequest, createPartyWithCredentialsResponse, err := CreatePartyWithCredentials(t, client)
		if err != nil {
			t.Fatalf("Failed to call CreatePartyWithCredentials %+v", err)
		}

		// Delete credential
		deletePrivateCredentialsRequest.Type =
			&credentials_service_type.DeletePrivateCredentialsRequest_TypeValue{
				TypeValue: &credentials_service_type.DeletePrivateCredentialsRequest_ByTypeValue{
					Value: createPartyWithCredentialsRequest.CreatePublicCredentialsRequests[0].Value,
				},
			}
		deleteCredentialsResponse, err = client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePrivateCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST, deleteCredentialsResponse.Status)
		}

		// valid request
		deletePrivateCredentialsRequest.Type =
			&credentials_service_type.DeletePrivateCredentialsRequest_TypeValue{
				TypeValue: &credentials_service_type.DeletePrivateCredentialsRequest_ByTypeValue{
					Value: createPartyWithCredentialsRequest.CreatePublicCredentialsRequests[0].Value,
					PrivateCredentialType: privatecredentialtype.PrivateCredentialType_BIOMETRIC_TOKEN,
				},
			}
		switch x := deletePrivateCredentialsRequest.Type.(type) {
		case *credentials_service_type.DeletePrivateCredentialsRequest_TypeValue:
			x.TypeValue.Value = createPartyWithCredentialsRequest.CreatePrivateCredentialsRequests[0].Value.GetStringValue()
			x.TypeValue.PartyId = createPartyWithCredentialsResponse.PartyId
		}
		t.Logf("Request %+v", deletePrivateCredentialsRequest)
		deleteCredentialsResponse, err = client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePrivateCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_SUCCESS, deleteCredentialsResponse.Status)
		}

		// and again to prove that it fails the second time
		deleteCredentialsResponse, err = client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePrivateCredentials %+v", err)
		case deleteCredentialsResponse.Status != credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS, deleteCredentialsResponse.Status)
		}

		return nil
	}, t)
}

func TestServer_DeletePrivateCredentialsById(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		deletePrivateCredentialsByIdRequest := &credentials_service_type.DeletePrivateCredentialsRequest{
			Type: &credentials_service_type.DeletePrivateCredentialsRequest_CredentialsId{
				CredentialsId: "",
			},
		}
		deleteCredentialsByIdResponse, err := client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsByIdRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call CreatePrivateCredentials %+v", err)
		case deleteCredentialsByIdResponse.Status != credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_INVALID_REQUEST, deleteCredentialsByIdResponse.Status)
		}

		// create Party and credentials
		_, createPartyWithCredentialsResponse, _ := CreatePartyWithCredentials(t, client)

		// Delete credential
		switch x := deletePrivateCredentialsByIdRequest.Type.(type) {
		case *credentials_service_type.DeletePrivateCredentialsRequest_CredentialsId:
			x.CredentialsId = createPartyWithCredentialsResponse.PrivateCredentialIds[0]
		}
		//credentials_service_type.DeleteCredentialsRequest_CredentialsId(deletePrivateCredentialsByIdRequest.Type).CredentialsId = createPartyWithCredentialsResponse.PrivateCredentialIds[0]
		deleteCredentialsByIdResponse, err = client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsByIdRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePrivateCredentials %+v", err)
		case deleteCredentialsByIdResponse.Status != credentials_service_type.DeleteCredentialsResponse_SUCCESS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_SUCCESS, deleteCredentialsByIdResponse.Status)
		}

		deleteCredentialsByIdResponse, err = client.DeletePrivateCredentials(context.Background(), deletePrivateCredentialsByIdRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call DeletePrivateCredentials %+v", err)
		case deleteCredentialsByIdResponse.Status != credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS:
			t.Fatalf("Invalid status, expected %s, got %s", credentials_service_type.DeleteCredentialsResponse_CREDENTIAL_NOT_EXISTS, deleteCredentialsByIdResponse.Status)
		}

		return nil
	}, t)
}

func TestServer_UpdatePartyStatus(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		// successful request
		request := h.CreatePartyRequest()
		createPartyResponse, err := client.CreateParty(context.Background(), request)
		if err != nil {
			t.Fatalf("Failed to call CreateParty: %+v", err)
		}
		t.Logf("createPartyResponse: %+v", createPartyResponse)
		validateCreateParty(t, createPartyResponse, err)

		in := &party_service_type.UpdatePartyStatusRequest{
			PartyId:     createPartyResponse.PartyId,
			PartyStatus: partystatus.PartyStatus_INACTIVE,
		}

		updatePartyResponse, err := client.UpdatePartyStatus(context.Background(), in)
		if err != nil {
			t.Fatalf("Failed to call UpdatePartyStatus: %+v", err)
		}
		t.Logf("updatePartyResponse: %+v", updatePartyResponse)

		retrievePartyResponse, err := client.RetrieveParty(context.Background(), &party_service_type.RetrievePartyRequest{
			PartyId: createPartyResponse.PartyId,
		})
		if err != nil {
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		}
		t.Logf("retrievePartyResponse: %+v", retrievePartyResponse)

		status := retrievePartyResponse.GetParty().PartyStatus

		if retrievePartyResponse.GetParty().PartyStatus != partystatus.PartyStatus_INACTIVE {
			t.Fatalf("Failed to retrieve valid status. Expected: %v, got %v", partystatus.PartyStatus_INACTIVE, status)
		}

		return nil
	}, t)
}

func TestServer_UpdatePartyName(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		request := h.CreatePartyRequest()
		createPartyResponse, err := client.CreateParty(context.Background(), request)
		if err != nil {
			t.Fatalf("Failed to call CreateParty: %+v", err)
		}
		t.Logf("createPartyResponse: %+v", createPartyResponse)
		validateCreateParty(t, createPartyResponse, err)

		firstname := uuid.New().String()
		lastname := uuid.New().String()

		in := &party_service_type.UpdatePartyNameRequest{
			PartyId: createPartyResponse.PartyId,
			Party: &party_service_type.UpdatePartyNameRequest_Person_{
				Person: &party_service_type.UpdatePartyNameRequest_Person{
					Name:      firstname + " " + lastname,
					FirstName: firstname,
					LastName:  lastname,
				},
			},
		}

		updatePartyResponse, err := client.UpdatePartyName(context.Background(), in)
		if err != nil {
			t.Fatalf("Failed to call UpdatePartyName: %+v", err)
		}
		t.Logf("updatePartyResponse: %+v", updatePartyResponse)

		if updatePartyResponse.Status != party_service_type.UpdatePartyNameResponse_SUCCESS {
			t.Fatalf("Failed to call UpdatePartyName: %+v", err)
		}

		retrievePartyResponse, err := client.RetrieveParty(context.Background(), &party_service_type.RetrievePartyRequest{
			PartyId: createPartyResponse.PartyId,
		})

		if err != nil {
			t.Fatalf("Failed to call RetrieveParty: %+v", err)
		}
		t.Logf("retrievePartyResponse: %+v", retrievePartyResponse)

		partyData := retrievePartyResponse.Party.PartyData

		if partyData.GetPerson().FirstName != firstname {
			t.Fatalf("Expected firstname %s got %s", firstname, partyData.GetPerson().FirstName)
		}

		if partyData.GetPerson().LastName != lastname {
			t.Fatalf("Expected lastname %s got %s", lastname, partyData.GetPerson().LastName)
		}

		return nil
	}, t)
}

func TestServer_UpdateInvalidPartyName(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {

		firstname := uuid.New().String()
		lastname := uuid.New().String()

		in := &party_service_type.UpdatePartyNameRequest{
			PartyId: uuid.New().String(),
			Party: &party_service_type.UpdatePartyNameRequest_Person_{
				Person: &party_service_type.UpdatePartyNameRequest_Person{
					Name:      firstname + " " + lastname,
					FirstName: firstname,
					LastName:  lastname,
				},
			},
		}

		response, err := client.UpdatePartyName(context.Background(), in)
		if err != nil {
			t.Fatalf("Expected UpdatePartyName to not fail: %+v", err)
		}

		if response.Status != party_service_type.UpdatePartyNameResponse_PARTY_NOT_EXISTS {
			t.Fatalf("Expected a response of %s got %s", party_service_type.UpdatePartyNameResponse_PARTY_NOT_EXISTS, response.Status)
		}

		return nil
	}, t)
}
