package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/tidwall/gjson"
)

type input struct {
	ToolDisplayName string `json:"tool_display_name,omitempty"`
	Message         string `json:"message"`
	UsernameField   string `json:"username_field,omitempty"`
	UsernameEnv     string `json:"username_env,omitempty"`
	PasswordField   string `json:"password_field,omitempty"`
	PasswordEnv     string `json:"password_env,omitempty"`
	Metadata        map[string]string
}

type sysPromptInput struct {
	Message   string            `json:"message,omitempty"`
	Fields    string            `json:"fields,omitempty"`
	Sensitive string            `json:"sensitive,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func main() {
	// Set up signal handler
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	in, err := getInput()
	if err != nil {
		fmt.Println("Input error: ", err)
		os.Exit(1)
	}

	username, password, err := getCredentials(ctx, in)
	if err != nil {
		fmt.Println("Error getting credentials:", err)
		os.Exit(1)
	}
	fmt.Printf(`{"env": {"%s": "%s", "%s": "%s"}}`, in.UsernameEnv, username, in.PasswordEnv, password)
}

func getCredentials(ctx context.Context, in input) (string, string, error) {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		fmt.Println("Error creating GPTScript client:", err)
		return "", "", fmt.Errorf("Error creating GPTScript client: %w", err)
	}
	defer client.Close()

	sysPromptIn, err := json.Marshal(sysPromptInput{
		Message:   in.Message,
		Fields:    strings.Join([]string{in.UsernameField, in.PasswordField}, ","),
		Sensitive: strconv.FormatBool(true),
		Metadata:  in.Metadata,
	})
	if err != nil {
		return "", "", fmt.Errorf("Error marshalling sys prompt input: %w", err)
	}

	run, err := client.Run(ctx, "sys.prompt", gptscript.Options{
		Input: string(sysPromptIn),
	})
	if err != nil {
		return "", "", fmt.Errorf("Error running GPTScript prompt: %w", err)
	}

	res, err := run.Text()
	if err != nil {
		return "", "", fmt.Errorf("Error getting GPTScript response: %w", err)
	}

	username := gjson.Get(res, in.UsernameField).String()
	password := gjson.Get(res, in.PasswordField).String()

	return username, password, nil

}

func getInput() (input, error) {
	if len(os.Args) != 2 {
		return input{}, errors.New("Missing input string")
	}

	inputStr := os.Args[1]

	var in input
	if err := json.Unmarshal([]byte(inputStr), &in); err != nil {
		return input{}, fmt.Errorf("Error parsing input JSON: %w", err)
	}

	// Helper function to trim and set default values
	cleanField := func(field, defaultValue string) string {
		field = strings.TrimSpace(field)
		if field == "" {
			return defaultValue
		}
		return field
	}

	in.ToolDisplayName = cleanField(in.ToolDisplayName, "Basic Auth Credential")
	in.UsernameField = cleanField(in.UsernameField, "username")
	in.PasswordField = cleanField(in.PasswordField, "password")

	// Set environment variables and validate
	var validEnvPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	cleanEnv := func(env, field string) (string, error) {
		env = strings.TrimSpace(env)
		if env == "" {
			return "", fmt.Errorf("Environment variable name for field %s must be set", field)
		}
		if !validEnvPattern.MatchString(env) {
			return "", fmt.Errorf("Invalid environment variable name: %s", env)
		}
		return env, nil
	}

	var err error
	in.UsernameEnv, err = cleanEnv(in.UsernameEnv, in.UsernameField)
	if err != nil {
		return input{}, err
	}
	in.PasswordEnv, err = cleanEnv(in.PasswordEnv, in.PasswordField)
	if err != nil {
		return input{}, err
	}

	in.Message = fmt.Sprintf("Enter your %s and %s", in.UsernameField, in.PasswordField)

	in.Metadata = map[string]string{
		"authType":        "basic",
		"toolContext":     "credential",
		"toolDisplayName": in.ToolDisplayName,
	}

	return in, nil
}
