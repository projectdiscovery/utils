package burpxml

import "fmt"

func (i Item) ToStrings(noReq, noResp bool) []string {
	arr := []string{
		i.Time,
		i.URL,
		i.Host.Name,
		i.Host.IP,
		i.Port,
		i.Protocol,
		i.Path,
		i.Extension,
	}

	if !noReq {
		arr = append(arr, i.Request.ToStrings()...)
	}

	arr = append(arr, []string{
		i.Status,
		i.ResponseLength,
		i.MimeType,
	}...)

	if !noResp {
		arr = append(arr, i.Response.ToStrings()...)
	}

	arr = append(arr, i.Comment)
	return arr
}

func (r Request) ToStrings() []string {
	if r.Body != "" {
		return []string{r.Body}
	}
	return []string{r.Base64Encoded, r.Raw}
}

func (r Response) ToStrings() []string {
	if r.Body != "" {
		return []string{r.Body}
	}
	return []string{r.Base64Encoded, r.Raw}
}

func (i Item) FlatString() string {
	return fmt.Sprintf(`%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s`,
		i.Time,
		i.URL,
		i.Host.Name,
		i.Host.IP,
		i.Port,
		i.Protocol,
		i.Path,
		i.Extension,
		i.Request.FlatString(),
		i.Status,
		i.ResponseLength,
		i.MimeType,
		i.Response.FlatString(),
		i.Comment,
	)
}

func (r Request) FlatString() string {
	if r.Body != "" {
		return r.Body
	}
	return fmt.Sprintf(`%s,%s`, r.Base64Encoded, r.Raw)
}

func (r Response) FlatString() string {
	if r.Body != "" {
		return r.Body
	}
	return fmt.Sprintf(`%s,%s`, r.Base64Encoded, r.Raw)
}

func (i Item) String() string {
	return fmt.Sprintf(`Item{
	Time	=	%s,
	Url	=	%s,
	Host	=	%s,
	IP	=	%s,
	Port	=	%s,
	Proto	=	%s,
	Path	=	%s,
	Ext	=	%s,
	%s,
	Status	=	%s,
	RespLen	=	%s,
	MIME	=	%s,
	%s,
	Comment	=	%s,
}`,
		i.Time,
		i.URL,
		i.Host.Name,
		i.Host.IP,
		i.Port,
		i.Protocol,
		i.Path,
		i.Extension,
		i.Request.String(),
		i.Status,
		i.ResponseLength,
		i.MimeType,
		i.Response.String(),
		i.Comment,
	)
}

func (r Request) String() string {
	s := "Request{\n"
	if r.Body != "" {
		s += fmt.Sprintf("	Body = %s,\n", r.Body)
	} else {
		s += fmt.Sprintf("	Base64	=	%s,\nBody	=	%s,\n", r.Base64Encoded, r.Raw)
	}
	s += "}"
	return s
}

func (r Response) String() string {
	s := "Response{\n"
	if r.Body != "" {
		s += fmt.Sprintf("	Body = %s,\n", r.Body)
	} else {
		s += fmt.Sprintf("	Base64	=	%s,\nBody	=	%s,\n", r.Base64Encoded, r.Raw)
	}
	s += "}"
	return s
}
