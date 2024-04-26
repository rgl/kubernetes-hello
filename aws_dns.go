package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func getAWSDNSZones(ctx context.Context) []nameValuePair {
	// bail when not running in aws.
	// see https://docs.aws.amazon.com/eks/latest/userguide/pod-id-how-it-works.html
	if os.Getenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE") == "" {
		return []nameValuePair{}
	}

	var zones nameValuePairs

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return []nameValuePair{{"ERROR", fmt.Sprintf("%v", err)}}
	}

	client := route53.NewFromConfig(cfg)

	response, err := client.ListHostedZones(ctx, nil)
	if err != nil {
		return []nameValuePair{{"ERROR", fmt.Sprintf("%v", err)}}
	}

	for _, hostedZone := range response.HostedZones {
		zone, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{Id: hostedZone.Id})
		if err != nil {
			return []nameValuePair{{"ERROR", fmt.Sprintf("%v", err)}}
		}
		zones = append(zones, nameValuePair{
			Name:  *zone.HostedZone.Name,
			Value: strings.Join(zone.DelegationSet.NameServers, "\n")})
	}

	sort.Sort(zones)

	return zones
}
