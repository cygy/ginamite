package location

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cygy/ginamite/common/config"
	"github.com/cygy/ginamite/common/queue"
)

// GetIPInfo : retrieve the location about an IP address.
func GetIPInfo(payload []byte, saveIPAddressDetailsFunc func(IPAddress, tokenID string, getIPAddressDetailsFunc func(IPAddress string) *IPAddressDetails)) {
	if saveIPAddressDetailsFunc == nil {
		return
	}

	p := queue.TaskIPLocation{}
	queue.UnmarshalPayload(payload, &p)

	saveIPAddressDetailsFunc(p.IPAddress, p.TokenID, GetIPAddressDetails)
}

// GetIPAddressDetails : returns the feo details of an IP address
func GetIPAddressDetails(IPAddress string) *IPAddressDetails {
	// Get the location.
	httpRequest, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s?access_key=%s", config.Main.IPStack.Host, IPAddress, config.Main.IPStack.Key), nil)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		logError(IPAddress, err)
		return nil
	}

	defer httpResponse.Body.Close()
	responseContent, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		logError(IPAddress, err)
		return nil
	}

	details := IPAddressDetails{}
	if err := json.Unmarshal(responseContent, &details); err != nil {
		logError(IPAddress, err)
		return nil
	}

	if len(details.Region) == 0 && len(details.CountryName) == 0 && len(details.CountryCode) == 0 {
		logError(IPAddress, errors.New("get no location from the API"))
		return nil
	}

	return &details
}
