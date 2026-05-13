package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	minPort = 1
	maxPort = 65535
)

var (
	errInvalidConnectionMapping = errors.New("invalid connection mapping")
	errInvalidSourcePort        = errors.New("invalid source port")
	errInvalidTargetPort        = errors.New("invalid target port")
	errPortOutOfRange           = fmt.Errorf("port out of range (must be %d-%d)", minPort, maxPort)
	errDuplicateSourcePort      = errors.New("duplicate source port")
)

func parseConnectionMappings(prefix string, env []string) ([]connectionMapping, error) {
	connectionMappings := []connectionMapping{}

	for _, envVar := range env {
		kv := strings.SplitN(envVar, "=", 2)

		if len(kv) != 2 {
			continue
		}

		if !strings.HasPrefix(kv[0], prefix) {
			continue
		}

		mapping, err := parseConnectionMapping(kv[1])
		if err != nil {
			return nil, err
		}

		connectionMappings = append(connectionMappings, mapping)
	}

	sourcePorts := []int{}

	for _, connectionMapping := range connectionMappings {
		if slices.Contains(sourcePorts, connectionMapping.SourcePort) {
			return nil, fmt.Errorf("%w: %d", errDuplicateSourcePort, connectionMapping.SourcePort)
		}

		sourcePorts = append(sourcePorts, connectionMapping.SourcePort)
	}

	return connectionMappings, nil
}

func parseConnectionMapping(value string) (connectionMapping, error) {
	original := value
	https := false

	end := 0
	for end < len(value) && isASCIILetter(value[end]) {
		end++
	}

	if end > 0 && end < len(value) && value[end] == ':' {
		candidate := value[end+1:]
		if strings.EqualFold(value[:end], "https") {
			https = true
			value = candidate
		} else if strings.Count(candidate, ":") == 2 {
			// Unknown protocol: strip only if doing so leaves a clean 3-part
			// mapping, otherwise leave the value alone so bad source ports still
			// surface as errInvalidSourcePort.
			value = candidate
		}
	}

	parts := strings.Split(value, ":")

	if len(parts) != 3 {
		return connectionMapping{}, fmt.Errorf("%w: %s (expected [https:]<source_port>:<target_host>:<target_port>)", errInvalidConnectionMapping, original)
	}

	sourcePort, err := strconv.Atoi(parts[0])
	if err != nil {
		return connectionMapping{}, fmt.Errorf("%w: %s", errInvalidSourcePort, parts[0])
	}
	if sourcePort < minPort || sourcePort > maxPort {
		return connectionMapping{}, fmt.Errorf("%w: %w: %d", errInvalidSourcePort, errPortOutOfRange, sourcePort)
	}

	targetPort, err := strconv.Atoi(parts[2])
	if err != nil {
		return connectionMapping{}, fmt.Errorf("%w: %s", errInvalidTargetPort, parts[2])
	}
	if targetPort < minPort || targetPort > maxPort {
		return connectionMapping{}, fmt.Errorf("%w: %w: %d", errInvalidTargetPort, errPortOutOfRange, targetPort)
	}

	return connectionMapping{
		SourcePort: sourcePort,
		TargetAddr: parts[1],
		TargetPort: targetPort,
		HTTPS:      https,
	}, nil
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
