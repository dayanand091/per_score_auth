package identity_server_test

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/wrsinc/gogenproto/identity/enums/addresstype"
	"github.com/wrsinc/gogenproto/identity/enums/partyownertype"
	"github.com/wrsinc/gogenproto/identity/enums/partyrelationshipstatus"
	"github.com/wrsinc/gogenproto/identity/enums/partyrelationshiptype"
	"github.com/wrsinc/gogenproto/identity/enums/partytype"
	"github.com/wrsinc/gogenproto/identity/enums/phonenumbertype"
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
	"github.com/wrsinc/identity/server"
	test_helpers "github.com/wrsinc/identity/testhelpers"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var once sync.Once

type fnTestFunction func(t *testing.T, client identity_service.IdentityClient) error

func startServer(t *testing.T) {
	once.Do(func() {
		t.Log("Starting server...")
		go server.StartServer("", "", "")
		time.Sleep(time.Millisecond * 3000)
	})
}

func testRunner(fn fnTestFunction, t *testing.T) error {
	startServer(t)

	conn, err := grpc.Dial(params.Port, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %+v", err)
	}
	defer conn.Close()
	client := identity_service.NewIdentityClient(conn)
	fn(t, client)

	return nil
}

func validateCreateParty(t *testing.T, response *party_service_type.CreatePartyResponse, err error) {
	// Validation function for Party
	switch {
	case err != nil:
		t.Fatalf("Failed to create Party %+v", err)
	case response.Status != party_service_type.CreatePartyResponse_SUCCESS:
		t.Fatalf("Invalid status, expected SUCCESS, got %s", response.Status)
	}
}

func validateCreatePartyWithCredentials(t *testing.T, request *party_service_type.CreatePartyWithCredentialsRequest, response *party_service_type.CreatePartyWithCredentialsResponse, err error) {

	switch {
	case err != nil:
		t.Fatalf("Failed to create Party %+v", err)
	case response.Status != party_service_type.CreatePartyWithCredentialsResponse_SUCCESS:
		t.Fatalf("Invalid status, expected SUCCESS, got %s", response.Status)
	case len(response.PublicCredentialIds) != len(request.CreatePublicCredentialsRequests):
		t.Fatalf("Invalid public credential count, expected %d, got %d", len(request.CreatePublicCredentialsRequests), len(response.PublicCredentialIds))
	case len(response.PrivateCredentialIds) != len(request.CreatePrivateCredentialsRequests):
		t.Fatalf("Invalid public credential count, expected %d, got %d", len(request.CreatePrivateCredentialsRequests), len(response.PrivateCredentialIds))
	}
	for _, id := range response.PublicCredentialIds {
		if id == "" {
			t.Fatal("Received empty publicCredentialId")
		}
	}
	for _, id := range response.PrivateCredentialIds {
		if id == "" {
			t.Fatal("Received empty privateCredentialId")
		}
	}
}

func validatePartyRelationship(t *testing.T, response *party_relationship_service_type.CreatePartyRelationshipResponse, status party_relationship_service_type.CreatePartyRelationshipResponse_Status, err error) {
	switch {
	case err != nil:
		t.Fatalf("Error response %+v", err)
	case response.Status != status:
		t.Fatalf("Invalid status, expected %s got %s", status, response.Status)
	}
}

func validateCreateCredentials(t *testing.T, response *credentials_service_type.CreateCredentialsResponse, status credentials_service_type.CreateCredentialsResponse_Status, err error) {
	// Validation function for CreateCredentials
	t.Logf("response %+v", response)
	switch {
	case err != nil:
		t.Fatalf("Failed to call CreatePublicCredentials: %+v", err)
	case response.Status != status:
		t.Fatalf("Invalid status, expected %s, got %s", status, response.Status)
	}
}

func CreatePartyRelationship(partyIdFrom, partyIdTo string) *party_relationship_service_type.CreatePartyRelationshipRequest {
	return &party_relationship_service_type.CreatePartyRelationshipRequest{
		PartyRelationshipType:   partyrelationshiptype.PartyRelationshipType_EMPLOYMENT,
		PartyIdFrom:             partyIdFrom,
		RoleTypeFrom:            roletype.RoleType_EMPLOYER,
		PartyIdTo:               partyIdTo,
		RoleTypeTo:              roletype.RoleType_EMPLOYEE,
		DateFrom:                &timestamp.Timestamp{},
		DateTo:                  &timestamp.Timestamp{},
		PartyRelationshipStatus: partyrelationshipstatus.PartyRelationshipStatus_ACTIVE,
	}
}

func CreatePartyWithCredentials(t *testing.T, client identity_service.IdentityClient) (*party_service_type.CreatePartyWithCredentialsRequest, *party_service_type.CreatePartyWithCredentialsResponse, error) {

	createPartyRequest := test_helpers.CreatePartyRequest()
	createPublicCredentialsRequests := make([]*credentials_service_type.CreatePublicCredentialsRequest, 1)
	createPublicCredentialsRequests[0] = test_helpers.CreatePublicCredentialRequest(
		publiccredentialtype.PublicCredentialType_USERNAME_PASSWORD, "", "", "", "")
	createPrivateCredentialsRequests := make([]*credentials_service_type.CreatePrivateCredentialsRequest, 1)
	createPrivateCredentialsRequests[0], _ = test_helpers.CreatePrivateCredentialRequest(
		privatecredentialtype.PrivateCredentialType_BIOMETRIC_TOKEN, "", "")
	createPartyWithCredentialsRequest := &party_service_type.CreatePartyWithCredentialsRequest{
		CreatePartyRequest:               createPartyRequest,
		CreatePublicCredentialsRequests:  createPublicCredentialsRequests,
		CreatePrivateCredentialsRequests: createPrivateCredentialsRequests,
	}
	createPartyWithCredentialsResponse, err := client.CreatePartyWithCredentials(context.Background(),
		createPartyWithCredentialsRequest)
	t.Logf("createPartyWithCredentialsRequest: %+v", createPartyWithCredentialsRequest)
	if err != nil {
		t.Fatalf("Failed to call CreateParty: %+v", err)
	}
	//t.Logf("createPartyWithCredentialsResponse: %+v", createPartyWithCredentialsResponse)

	return createPartyWithCredentialsRequest, createPartyWithCredentialsResponse, err
}

func CreateParty(t *testing.T, client identity_service.IdentityClient, request *party_service_type.CreatePartyRequest) *party_service_type.CreatePartyResponse {
	createPartyResponse, err := client.CreateParty(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to call CreateParty: %+v", err)
	}
	t.Logf("CreatePartyResponse: %+v", createPartyResponse)
	validateCreateParty(t, createPartyResponse, err)

	return createPartyResponse
}

func CreateRole(t *testing.T, client identity_service.IdentityClient, request *role_service_type.CreateRoleRequest) *role_service_type.CreateRoleResponse {
	createRoleResponse, err := client.CreateRole(context.Background(), request)

	switch {
	case err != nil:
		t.Fatalf("Error response %+v", err)
	case createRoleResponse.Status != role_service_type.CreateRoleResponse_SUCCESS:
		t.Fatalf("Invalid status, expected SUCCESS got %s", createRoleResponse.Status)
	}

	t.Logf("CreateRoleResponse: %+v", createRoleResponse)

	return createRoleResponse
}

func CreateOrganizationPartyRequest(tenantId, organizationName string) *party_service_type.CreatePartyRequest {
	return &party_service_type.CreatePartyRequest{
		TenantId:       tenantId,
		PartyType:      partytype.PartyType_ORGANIZATION,
		PartyOwnerType: partyownertype.PartyOwnerType_FIRST_PARTY,
		PartyOwner:     "WRS",
		DateFrom:       &timestamp.Timestamp{},
		DateTo:         &timestamp.Timestamp{},
		PartyData: &common_service_type.PartyData{
			PartyData: &common_service_type.PartyData_Organization{
				Organization: &common_service_type.Organization{
					Name: organizationName,
					BusinessPhoneNumber: &common_service_type.PhoneNumber{
						PhoneNumberKey:  uuid.New().String(),
						PhoneNumberType: phonenumbertype.PhoneNumberType_BUSINESS,
						CountryCode:     "US",
						Number:          "555-555-5556",
					},
					ContactPhoneNumber: &common_service_type.PhoneNumber{
						PhoneNumberKey:  uuid.New().String(),
						PhoneNumberType: phonenumbertype.PhoneNumberType_BUSINESS,
						CountryCode:     "US",
						Number:          "555-555-5555",
					},
					Address: &common_service_type.Address{
						AddressKey:  uuid.New().String(),
						AddressType: addresstype.AddressType_HOME,
						Nickname:    "abc",
						StreetLines: []string{"1 1st street"},
						Locality:    "Austin",
						Region:      "Texas",
						PostalCode:  "78777",
						Country:     "USA",
					},
					ContactEmailAddress: &common_service_type.EmailAddress{
						EmailKey: uuid.New().String(),
						Email:    uuid.New().String() + "@westfield.com",
						Verified: true,
					},
				},
			},
		},
	}
}
