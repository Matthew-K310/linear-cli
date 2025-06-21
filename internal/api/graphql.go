package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const GraphQLEndpoint = "https://api.linear.app/graphql" // Define the endpoint here

// GraphQLRequest represents the structure for a GraphQL request body.
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents the structure of a GraphQL response body.
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

// MakeGraphQLRequest sends a GraphQL query to the Linear API endpoint.
// It requires the API key, the query string, and optional variables.
// It returns the raw JSON data from the 'data' field or an error.
func MakeGraphQLRequest(
	apiKey string,
	query string,
	variables map[string]interface{},
) ([]byte, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("APIKey is not provided for the request")
	}

	graphQLReqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(graphQLReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request body: %w", err)
	}

	req, err := http.NewRequest("POST", GraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", apiKey) // Use the provided apiKey

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response body: %w", err)
	}

	// Check for non-OK status codes (e.g., 401, 403, 400, 500)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"GraphQL request returned non-OK status code %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var graphQLResp GraphQLResponse
	err = json.Unmarshal(respBody, &graphQLResp)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse GraphQL response JSON: %w - response body: %s",
			err,
			string(respBody),
		)
	}

	// Check for GraphQL errors returned in the response body
	// Note: Some APIs return a 200 OK even with GraphQL errors.
	if len(graphQLResp.Errors) > 0 && string(graphQLResp.Errors) != "null" {
		// You might want to parse the errors more carefully here
		return nil, fmt.Errorf("GraphQL errors in response: %s", string(graphQLResp.Errors))
	}

	return graphQLResp.Data, nil
}
