package burpxml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

type XMLParseOptions struct {
	DecodeBase64 bool
}

func ParseXML(r io.Reader, opts XMLParseOptions) (*Items, error) {
	var items Items
	if err := xml.NewDecoder(r).Decode(&items); err != nil {
		return nil, fmt.Errorf("xml decode: %w", err)
	}

	if !opts.DecodeBase64 {
		return &items, nil
	}

	for i := range items.Items {
		if err := decodeItemBodies(&items.Items[i]); err != nil {
			return nil, err
		}
	}

	return &items, nil
}

func decodeItemBodies(item *Item) error {
	decoded, err := decodeBase64Field(item.Request.Base64Encoded, item.Request.Raw)
	if err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	item.Request.Body = decoded

	decoded, err = decodeBase64Field(item.Response.Base64Encoded, item.Response.Raw)
	if err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	item.Response.Body = decoded

	return nil
}

func decodeBase64Field(base64Flag, raw string) (string, error) {
	if base64Flag == "" {
		return raw, nil
	}

	isBase64, err := strconv.ParseBool(base64Flag)
	if err != nil {
		return "", fmt.Errorf("parse base64 flag: %w", err)
	}

	if !isBase64 {
		return raw, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	return string(decoded), nil
}
