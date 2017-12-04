package identity_server_test

import (
	"testing"

	"github.com/google/uuid"
	identity_service "github.com/wrsinc/gogenproto/identity/service"
	"github.com/wrsinc/gogenproto/identity/service/party"
	"golang.org/x/net/context"
)

func TestServer_CreateOrganizationAndUpdateName(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		tenantId := uuid.New().String()
		name := uuid.New().String()

		req := CreateOrganizationPartyRequest(tenantId, name)

		createPartyResponse := CreateParty(t, client, req)

		newName := uuid.New().String()
		in := &party.UpdatePartyNameRequest{
			PartyId: createPartyResponse.PartyId,
			Party: &party.UpdatePartyNameRequest_Organization_{
				Organization: &party.UpdatePartyNameRequest_Organization{
					Name: newName,
				},
			},
		}

		updatePartyResponse, err := client.UpdatePartyName(context.Background(), in)
		if err != nil {
			t.Fatalf("Failed to call UpdatePartyName: %+v", err)
		}
		t.Logf("updatePartyResponse: %+v", updatePartyResponse)

		if updatePartyResponse.Status != party.UpdatePartyNameResponse_SUCCESS {
			t.Fatalf("Failed to call UpdatePartyName: %+v", err)
		}

		retrievePartyRequest := &party.RetrievePartyRequest{
			PartyId: createPartyResponse.PartyId,
		}

		retrievePartyResponse, _ := client.RetrieveParty(context.Background(), retrievePartyRequest)

		if retrievePartyResponse.Party.PartyId != createPartyResponse.PartyId {
			t.Fatalf("Expected party id: %s got %s.", createPartyResponse.PartyId, retrievePartyResponse.Party.PartyId)
		}

		partyData := retrievePartyResponse.Party.GetPartyData()

		if partyData.GetOrganization().Name != newName {
			t.Fatalf("Expected org name %s got %s.", newName, partyData.GetOrganization().Name)
		}

		return nil
	}, t)
}
