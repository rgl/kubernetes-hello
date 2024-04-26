package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
)

func getAzureDNSZones(ctx context.Context) []nameValuePair {
	// bail when not running in azure.
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return []nameValuePair{}
	}

	// Azure AD Workload Identity webhook will inject the following environment variables:
	//	AZURE_AUTHORITY_HOST		the AAD authority hostname
	//	AZURE_CLIENT_ID				the client_id set in the service account annotation
	//	AZURE_FEDERATED_TOKEN_FILE	the workload identity service account token path
	//	AZURE_TENANT_ID				the tenant_id set in the service account annotation
	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tokenPath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")

	getTokenAssertion := func(ctx context.Context) (string, error) {
		token, err := os.ReadFile(tokenPath)
		if err != nil {
			return "", err
		}
		return string(token), nil
	}

	// see https://github.com/Azure/azure-workload-identity/tree/main/examples/msal-go
	credential, err := azidentity.NewClientAssertionCredential(tenantID, clientID, getTokenAssertion, nil)
	if err != nil {
		return []nameValuePair{{"ERROR", fmt.Sprintf("%v", err)}}
	}

	client, err := armdns.NewZonesClient(subscriptionID, credential, nil)
	if err != nil {
		return []nameValuePair{{"ERROR", fmt.Sprintf("%v", err)}}
	}

	var zones nameValuePairs

	for pager := client.NewListPager(nil); pager.More(); {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return []nameValuePair{{"ERROR", fmt.Sprintf("%v", err)}}
		}
		for _, zone := range page.Value {
			var sb strings.Builder
			for i, ns := range zone.Properties.NameServers {
				if i > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(*ns)
			}
			zones = append(zones, nameValuePair{*zone.Name, sb.String()})
		}
	}

	sort.Sort(zones)

	return zones
}
