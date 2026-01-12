package burp

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
)

type Request struct {
	Base64Encoded string `xml:"base64,attr" json:"base64_encoded,omitempty"`
	Raw           string `xml:",chardata" json:"raw,omitempty"`
	Body          string `xml:"-" json:"body,omitempty"`
}

type Response struct {
	Base64Encoded string `xml:"base64,attr" json:"base64_encoded,omitempty"`
	Raw           string `xml:",chardata" json:"raw,omitempty"`
	Body          string `xml:"-" json:"body,omitempty"`
}

type Host struct {
	IP   string `xml:"ip,attr" json:"ip,omitempty"`
	Name string `xml:",chardata" json:"name,omitempty"`
}

type Item struct {
	Time           string   `xml:"time" json:"time,omitempty"`
	URL            string   `xml:"url" json:"url,omitempty"`
	Host           Host     `xml:"host" json:"host,omitzero"`
	Port           string   `xml:"port" json:"port,omitempty"`
	Protocol       string   `xml:"protocol" json:"protocol,omitempty"`
	Path           string   `xml:"path" json:"path,omitempty"`
	Extension      string   `xml:"extension" json:"extension,omitempty"`
	Request        Request  `xml:"request" json:"request,omitzero"`
	Status         string   `xml:"status" json:"status,omitempty"`
	ResponseLength string   `xml:"responselength" json:"response_length,omitempty"`
	MimeType       string   `xml:"mimetype" json:"mime_type,omitempty"`
	Response       Response `xml:"response" json:"response,omitzero"`
	Comment        string   `xml:"comment" json:"comment,omitempty"`
}

type Items struct {
	Items []Item `xml:"item" json:"items,omitempty"`
}

func (items *Items) ToJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

type CSVOptions struct {
	ExcludeRequest  bool
	ExcludeResponse bool
}

func (items *Items) ToCSV(w io.Writer, opts CSVOptions) error {
	enc := csv.NewWriter(w)
	defer enc.Flush()

	for _, item := range items.Items {
		record := item.toRecord(opts)
		if err := enc.Write(record); err != nil {
			return fmt.Errorf("csv write: %w", err)
		}
	}

	return enc.Error()
}

func (item *Item) toRecord(opts CSVOptions) []string {
	record := []string{
		item.Time,
		item.URL,
		item.Host.Name,
		item.Host.IP,
		item.Port,
		item.Protocol,
		item.Path,
		item.Extension,
	}

	if !opts.ExcludeRequest {
		record = append(record, item.Request.content())
	}

	record = append(record, item.Status, item.ResponseLength, item.MimeType)

	if !opts.ExcludeResponse {
		record = append(record, item.Response.content())
	}

	record = append(record, item.Comment)
	return record
}

func (r *Request) content() string {
	if r.Body != "" {
		return r.Body
	}
	return r.Raw
}

func (r *Response) content() string {
	if r.Body != "" {
		return r.Body
	}
	return r.Raw
}
