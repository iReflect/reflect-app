package google

import (
	"encoding/json"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/script/v1"
	"io/ioutil"
	"log"
	"net/http"
)

// AppScriptExecutor ...
type AppScriptExecutor struct {
	ScriptID        string
	CredentialsFile string
	client          *http.Client
}

// getClient Returns a Google oauth client to run script functions with.
func (a *AppScriptExecutor) getClient(ctx context.Context, scopes ...string) (*http.Client, error) {
	if a.client == nil {
		data, err := ioutil.ReadFile(a.CredentialsFile)
		if err != nil {
			return nil, err
		}

		type credentialsFile struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RefreshToken string `json:"refresh_token"`
		}

		var f credentialsFile

		if err := json.Unmarshal(data, &f); err != nil {
			return nil, err
		}

		cfg := &oauth2.Config{
			ClientID:     f.ClientID,
			ClientSecret: f.ClientSecret,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		}
		tok := &oauth2.Token{RefreshToken: f.RefreshToken}

		a.client = oauth2.NewClient(ctx, cfg.TokenSource(ctx, tok))
	}

	return a.client, nil
}

// Run Executes a remote Apps script with given scriptID, functionName & parameters.
// parameters should be passed in correct order and types should match the expected in Apps script function.
func (a *AppScriptExecutor) Run(functionName string, parameters ...interface{}) (googleapi.RawMessage, error) {

	client, err := a.getClient(context.Background(), script.SpreadsheetsScope)
	if err != nil {
		log.Println("Unable to retrieve Client", err)
		return nil, err
	}

	// Generate a service object.
	srv, err := script.New(client)
	if err != nil {
		log.Println("Unable to retrieve script Client", err)
		return nil, err
	}

	// Create an execution request object.
	req := script.ExecutionRequest{Function: functionName, Parameters: parameters}
	log.Println("Executing Function ", functionName, " with params: ", parameters)

	// Make the API request.
	resp, err := srv.Scripts.Run(a.ScriptID, &req).Do()
	if err != nil {
		// The API encountered a problem before the script started executing.
		log.Println("Unable to execute Apps Script function.", err)
		return nil, err
	}

	if resp.Error != nil {
		// The API executed, but the script returned an error.

		log.Println("Script error message: ", resp.Error.Message)
		for index, trace := range resp.Error.Details {
			log.Println("Trace", index, string(trace))
		}

		return nil, err
	}

	return resp.Response, nil
}
