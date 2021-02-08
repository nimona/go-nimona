// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package object

type (
	Certificate struct {
		Metadata Metadata `nimona:"metadata:m,omitempty"`
		Nonce    string   `nimona:"nonce:s,omitempty"`
		Created  string   `nimona:"created:s,omitempty"`
		Expires  string   `nimona:"expires:s,omitempty"`
	}
	CertificateRequest struct {
		Metadata               Metadata `nimona:"metadata:m,omitempty"`
		ApplicationName        string   `nimona:"applicationName:s,omitempty"`
		ApplicationDescription string   `nimona:"applicationDescription:s,omitempty"`
		ApplicationURL         string   `nimona:"applicationURL:s,omitempty"`
		Subject                string   `nimona:"subject:s,omitempty"`
		Resources              []string `nimona:"resources:as,omitempty"`
		Actions                []string `nimona:"actions:as,omitempty"`
		Nonce                  string   `nimona:"nonce:s,omitempty"`
	}
)

func (e *Certificate) Type() string {
	return "nimona.io/Certificate"
}

func (e Certificate) ToObject() *Object {
	r := &Object{
		Type:     "nimona.io/Certificate",
		Metadata: e.Metadata,
		Data:     Map{},
	}
	// else
	// r.Data["nonce"] = String(e.Nonce)
	r.Data["nonce"] = String(e.Nonce)
	// else
	// r.Data["created"] = String(e.Created)
	r.Data["created"] = String(e.Created)
	// else
	// r.Data["expires"] = String(e.Expires)
	r.Data["expires"] = String(e.Expires)
	return r
}

func (e *Certificate) FromObject(o *Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["nonce"]; ok {
		if t, ok := v.(String); ok {
			e.Nonce = string(t)
		}
	}
	if v, ok := o.Data["created"]; ok {
		if t, ok := v.(String); ok {
			e.Created = string(t)
		}
	}
	if v, ok := o.Data["expires"]; ok {
		if t, ok := v.(String); ok {
			e.Expires = string(t)
		}
	}
	return nil
}

func (e *CertificateRequest) Type() string {
	return "nimona.io/CertificateRequest"
}

func (e CertificateRequest) ToObject() *Object {
	r := &Object{
		Type:     "nimona.io/CertificateRequest",
		Metadata: e.Metadata,
		Data:     Map{},
	}
	// else
	// r.Data["applicationName"] = String(e.ApplicationName)
	r.Data["applicationName"] = String(e.ApplicationName)
	// else
	// r.Data["applicationDescription"] = String(e.ApplicationDescription)
	r.Data["applicationDescription"] = String(e.ApplicationDescription)
	// else
	// r.Data["applicationURL"] = String(e.ApplicationURL)
	r.Data["applicationURL"] = String(e.ApplicationURL)
	// else
	// r.Data["subject"] = String(e.Subject)
	r.Data["subject"] = String(e.Subject)
	// if $member.IsRepeated
	if len(e.Resources) > 0 {
		// else
		// r.Data["resources"] = ToStringArray(e.Resources)
		rv := make(StringArray, len(e.Resources))
		for i, iv := range e.Resources {
			rv[i] = String(iv)
		}
		r.Data["resources"] = rv
	}
	// if $member.IsRepeated
	if len(e.Actions) > 0 {
		// else
		// r.Data["actions"] = ToStringArray(e.Actions)
		rv := make(StringArray, len(e.Actions))
		for i, iv := range e.Actions {
			rv[i] = String(iv)
		}
		r.Data["actions"] = rv
	}
	// else
	// r.Data["nonce"] = String(e.Nonce)
	r.Data["nonce"] = String(e.Nonce)
	return r
}

func (e *CertificateRequest) FromObject(o *Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["applicationName"]; ok {
		if t, ok := v.(String); ok {
			e.ApplicationName = string(t)
		}
	}
	if v, ok := o.Data["applicationDescription"]; ok {
		if t, ok := v.(String); ok {
			e.ApplicationDescription = string(t)
		}
	}
	if v, ok := o.Data["applicationURL"]; ok {
		if t, ok := v.(String); ok {
			e.ApplicationURL = string(t)
		}
	}
	if v, ok := o.Data["subject"]; ok {
		if t, ok := v.(String); ok {
			e.Subject = string(t)
		}
	}
	if v, ok := o.Data["resources"]; ok {
		if t, ok := v.(StringArray); ok {
			rv := make([]string, len(t))
			for i, iv := range t {
				rv[i] = string(iv)
			}
			e.Resources = rv
		}
	}
	if v, ok := o.Data["actions"]; ok {
		if t, ok := v.(StringArray); ok {
			rv := make([]string, len(t))
			for i, iv := range t {
				rv[i] = string(iv)
			}
			e.Actions = rv
		}
	}
	if v, ok := o.Data["nonce"]; ok {
		if t, ok := v.(String); ok {
			e.Nonce = string(t)
		}
	}
	return nil
}
