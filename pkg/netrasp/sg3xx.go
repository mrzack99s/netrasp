package netrasp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// sg3xx is the Netrasp driver for Cisco IOS.
type sg3xx struct {
	Connection connection
}

// Close connection to device.
func (i sg3xx) Close(ctx context.Context) error {
	i.Connection.Close(ctx)

	return nil
}

// Configure device.
func (i sg3xx) Configure(ctx context.Context, commands []string) (ConfigResult, error) {
	var result ConfigResult
	_, err := i.Run(ctx, "configure terminal")
	if err != nil {
		return result, fmt.Errorf("unable to enter config mode: %w", err)
	}
	for _, command := range commands {
		output, err := i.Run(ctx, command)
		configCommand := ConfigCommand{Command: command, Output: output}
		result.ConfigCommands = append(result.ConfigCommands, configCommand)
		if err != nil {
			return result, fmt.Errorf("unable to run command '%s': %w", command, err)
		}
	}
	_, err = i.Run(ctx, "end")
	if err != nil {
		return result, fmt.Errorf("unable to exit from config mode: %w", err)
	}

	return result, nil
}

// SaveConfig
func (i sg3xx) SaveConfig(ctx context.Context) error {
	_, err := i.RunUntil(ctx, "write", enablePrompt)
	if err != nil {
		return err
	}
	_, err = i.Run(ctx, "Y")

	if err != nil {
		return err
	}

	return nil
}

// Dial opens a connection to a device.
func (i sg3xx) Dial(ctx context.Context) error {
	commands := []string{"terminal width 511"}

	return establishConnection(ctx, i, i.Connection, i.basePrompt(), commands)
}

// Enable elevates privileges.
func (i sg3xx) Enable(ctx context.Context) error {
	_, err := i.RunUntil(ctx, "enable", enablePrompt)
	if err != nil {
		return err
	}
	host := i.Connection.GetHost()
	_, err = i.Run(ctx, host.enableSecret)

	if err != nil {
		return err
	}

	return nil
}

// Run executes a command on a device.
func (i sg3xx) Run(ctx context.Context, command string) (string, error) {
	output, err := i.RunUntil(ctx, command, i.basePrompt())
	if err != nil {
		return "", err
	}

	output = strings.ReplaceAll(output, "\r\n", "\n")
	lines := strings.Split(output, "\n")
	result := ""

	for i := 1; i < len(lines)-1; i++ {
		result += lines[i] + "\n"
	}

	return result, nil
}

// RunUntil executes a command and reads until the provided prompt.
func (i sg3xx) RunUntil(ctx context.Context, command string, prompt *regexp.Regexp) (string, error) {
	err := i.Connection.Send(ctx, command)
	if err != nil {
		return "", fmt.Errorf("unable to send command to device: %w", err)
	}

	reader := i.Connection.Recv(ctx)
	output, err := readUntilPrompt(ctx, reader, prompt)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (i sg3xx) basePrompt() *regexp.Regexp {
	return generalPrompt
}
