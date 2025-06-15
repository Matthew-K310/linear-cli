package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var APIKey string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf(
			"Warning: Error loading .env file: %v. Proceeding assuming env vars are set directly.\n",
			err,
		)
	}

	APIKey = os.Getenv("API_KEY")

	if APIKey == "" {
		log.Fatal("Error: API_KEY not set in .env file or environment variables.")
	}

	log.Println("APIKey loaded successfully from environment.")
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors json.RawMessage `json:"errors"`
}

func makeGraphQLRequest(
	apiURL string,
	query string,
	variables map[string]interface{},
) ([]byte, error) {
	if APIKey == "" {
		return nil, fmt.Errorf("APIKey is not set, cannot make authenticated request")
	}

	graphQLReqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(graphQLReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request body: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Header.Set("Authorization", APIKey)

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

	if len(graphQLResp.Errors) > 0 && string(graphQLResp.Errors) != "null" {
		return nil, fmt.Errorf("GraphQL errors in response: %s", string(graphQLResp.Errors))
	}

	return graphQLResp.Data, nil
}

func main() {
	graphQLEndpoint := "https://api.linear.app/graphql"

	query := `
	query Me {
		viewer {
			id
			name
			email
		}
	}
	`

	variables := map[string]interface{}{}

	fmt.Println("Making GraphQL request with API key...")

	// The 'undefined: makeGraphQLRequest' error occurred here because the function was not defined above.
	data, err := makeGraphQLRequest(graphQLEndpoint, query, variables)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making GraphQL request: %v\n", err)
		return
	}

	fmt.Println("GraphQL request successful!")
	fmt.Println("Response Data:")
	fmt.Println(string(data))

	type Viewer struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type ViewerData struct {
		Viewer Viewer `json:"viewer"`
	}

	var viewerData ViewerData
	if err := json.Unmarshal(data, &viewerData); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling viewer data: %v\n", err)
	} else {
		fmt.Printf("Unmarshalled Viewer ID: %s\n", viewerData.Viewer.ID)
		fmt.Printf("Unmarshalled Viewer Name: %s\n", viewerData.Viewer.Name)
	}
}
