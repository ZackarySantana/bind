package main

import (
	"fmt"
	"os"
	"time"

	"github.com/posthog/posthog-go"
)

func main() {
	client, _ := posthog.NewWithConfig(os.Getenv("POSTHOG_API_KEY"), posthog.Config{Endpoint: os.Getenv("POSTHOG_API_HOST")})
	defer client.Close()

	isMyFlagEnabledForUser, err := client.GetFeatureFlag(posthog.FeatureFlagPayload{
		Key:        "my-flag",
		DistinctId: "distinct-id",
	})
	fmt.Println("featureflag value:", isMyFlagEnabledForUser, "error:", err)

	err = client.Enqueue(posthog.Capture{
		Type:       "capture",
		DistinctId: "distinct-id",
		Event:      "my event",
		Properties: posthog.NewProperties().Set("property1", 123).Set("property2", "hello"),
	})
	fmt.Println("Enqueue error:", err)

	time.Sleep(2 * time.Second) // wait for events to be sent
}
