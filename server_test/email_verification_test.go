package identity_server_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	identity_service "github.com/wrsinc/gogenproto/identity/service"
	"github.com/wrsinc/gogenproto/identity/service/party"
	"github.com/wrsinc/identity/manager"
	test_helpers "github.com/wrsinc/identity/testhelpers"
	"golang.org/x/net/context"
)

func TestServer_GeneratePersonEmailValidationLinks(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {

		createPartyRequest := test_helpers.CreatePartyRequest()
		createPartyResponse := CreateParty(t, client, createPartyRequest)

		emailKey := createPartyRequest.PartyData.GetPerson().Emails[0].EmailKey

		generatePersonEmailValidationLinksRequest := &party.GeneratePersonEmailValidationLinksRequest{PartyId: createPartyResponse.PartyId, EmailKey: emailKey}

		response, err := client.GeneratePersonEmailValidationLinks(context.Background(), generatePersonEmailValidationLinksRequest)
		switch {
		case err != nil:
			t.Fatalf("Failed to call GeneratePersonEmailValidationLinks: %+v", err)
		case response.Status != party.GeneratePersonEmailValidationLinksResponse_SUCCESS:
			t.Fatalf("Invalid Status, expected %s, got, %s", party.GeneratePersonEmailValidationLinksResponse_SUCCESS, response.Status)
		}

		if len(response.ValidationToken) == 0 {
			t.Fatalf("Invalid validation token, expected length greater than 0")
		}

		tokenBytes, err := base64.URLEncoding.DecodeString(response.ValidationToken)
		if err != nil {
			t.Fatalf("validation token not base64 encoded %+v", err)
		}

		var token party.SingleUseToken

		if err = proto.Unmarshal(tokenBytes, &token); err != nil {
			t.Fatalf("validation token not a protobuf %+v", err)
		}

		// valid party id
		if token.PartyId != createPartyResponse.PartyId {
			t.Fatalf("Invalid party id, expected %s, got, %s", createPartyResponse.PartyId, token.PartyId)
		}

		// valid email key
		if token.GetEmailVerification().EmailKey != emailKey {
			t.Fatalf("Invalid email key, expected %s, got, %s", emailKey, token.GetEmailVerification().EmailKey)
		}

		tokenExpireTime := time.Unix(token.ExpirationDate.Seconds, 0)
		timeDifference := time.Now().Add(24 * time.Hour).Sub(tokenExpireTime)

		if timeDifference >= 10*time.Second {
			t.Fatalf("expected a token expiration time difference of less than 10 seconds got %+v", timeDifference)
		}

		return nil
	}, t)
}

func TestServer_GeneratePersonEmailInvalidPartyId(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {

		pr := &party.GeneratePersonEmailValidationLinksRequest{PartyId: "INVALID", EmailKey: "tkuchlein@westfield.com"}

		resp, err := client.GeneratePersonEmailValidationLinks(context.Background(), pr)
		switch {
		case err != nil:
			t.Fatalf("Failed to call GeneratePersonEmailValidationLinks: %+v", err)
		case resp.Status != party.GeneratePersonEmailValidationLinksResponse_PARTY_NOT_EXISTS:
			t.Fatalf("Invalid Status, expected %s, got, %s", party.GeneratePersonEmailValidationLinksResponse_PARTY_NOT_EXISTS, resp.Status)
		}

		return nil
	}, t)
}

func TestServer_VerifyPersonEmail(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		createPartyRequest := test_helpers.CreatePartyRequest()
		createPartyResponse := CreateParty(t, client, createPartyRequest)

		emailKey := createPartyRequest.PartyData.GetPerson().Emails[0].EmailKey

		partyRequest := &party.GeneratePersonEmailValidationLinksRequest{PartyId: createPartyResponse.PartyId, EmailKey: emailKey}

		generatePersonEmailValidationLinksResponse, _ := client.GeneratePersonEmailValidationLinks(context.Background(), partyRequest)

		request := &party.VerifyPersonEmailRequest{ValidationToken: generatePersonEmailValidationLinksResponse.ValidationToken}
		verifyPersonEmailResponse, error := client.VerifyPersonEmail(context.Background(), request)

		switch {
		case error != nil:
			t.Fatalf("Failed to call VerifyPerson: %+v", error)
		case verifyPersonEmailResponse.Status != party.VerifyPersonEmailResponse_SUCCESS:
			t.Fatalf("Invalid status, expected to be %s, got %s", party.VerifyPersonEmailResponse_SUCCESS, verifyPersonEmailResponse.Status)
		}

		retrievePartyRequest := &party.RetrievePartyRequest{PartyId: createPartyResponse.PartyId}
		party, _ := client.RetrieveParty(context.Background(), retrievePartyRequest)

		person := party.Party.PartyData.GetPerson()

		switch {
		case person.Emails[0].Verified != true:
			t.Fatalf("Email.Verfied is still false")
		}

		return nil
	}, t)
}

func TestServer_VerifyPersonEmailBadPartyId(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		testUUID := uuid.New().String()
		token, _ := managers.CreateEmailValidationToken(testUUID, "test@westfield.com", time.Now().Add(1*time.Hour))

		request := &party.VerifyPersonEmailRequest{ValidationToken: token}
		response, error := client.VerifyPersonEmail(context.Background(), request)

		switch {
		case error != nil:
			t.Fatalf("An error occured during VerifyPersonEmail when called with a partyId that does not exist")
		case response.Status != party.VerifyPersonEmailResponse_PARTY_NOT_EXISTS:
			t.Fatalf("Invalid status, expected to be %s but got %s", party.VerifyPersonEmailResponse_PARTY_NOT_EXISTS, response.Status)
		}

		return nil

	}, t)
}

func TestServer_VerifyPersonInvalidToken(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		request := &party.VerifyPersonEmailRequest{ValidationToken: "1234"}
		response, error := client.VerifyPersonEmail(context.Background(), request)

		switch {
		case response.Status != party.VerifyPersonEmailResponse_TOKEN_NOT_VALID:
			t.Fatalf("Invalid token check has failed")
		case error != nil:
			t.Fatalf("An error was thrown during validation of the token")
		}

		return nil
	}, t)
}

func TestServer_VerifyPersonExpiredToken(t *testing.T) {
	testRunner(func(t *testing.T, client identity_service.IdentityClient) error {
		testUUID := uuid.New().String()
		token, _ := managers.CreateEmailValidationToken(testUUID, "test@westfield.com", time.Now().Add(-1*time.Hour))

		request := &party.VerifyPersonEmailRequest{ValidationToken: token}

		response, error := client.VerifyPersonEmail(context.Background(), request)

		switch {
		case response.Status != party.VerifyPersonEmailResponse_TOKEN_EXPIRED:
			t.Fatalf("Expired token was not caught")
		case error != nil:
			t.Fatalf("An error was thrown during validation of the token")
		}
		return nil
	}, t)
}
