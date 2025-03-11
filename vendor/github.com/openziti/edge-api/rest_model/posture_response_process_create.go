// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// PostureResponseProcessCreate posture response process create
//
// swagger:model postureResponseProcessCreate
type PostureResponseProcessCreate struct {
	idField *string

	// hash
	Hash string `json:"hash,omitempty"`

	// is running
	IsRunning bool `json:"isRunning,omitempty"`

	// path
	Path string `json:"path,omitempty"`

	// signer fingerprints
	SignerFingerprints []string `json:"signerFingerprints"`
}

// ID gets the id of this subtype
func (m *PostureResponseProcessCreate) ID() *string {
	return m.idField
}

// SetID sets the id of this subtype
func (m *PostureResponseProcessCreate) SetID(val *string) {
	m.idField = val
}

// TypeID gets the type Id of this subtype
func (m *PostureResponseProcessCreate) TypeID() PostureCheckType {
	return "PROCESS"
}

// SetTypeID sets the type Id of this subtype
func (m *PostureResponseProcessCreate) SetTypeID(val PostureCheckType) {
}

// UnmarshalJSON unmarshals this object with a polymorphic type from a JSON structure
func (m *PostureResponseProcessCreate) UnmarshalJSON(raw []byte) error {
	var data struct {

		// hash
		Hash string `json:"hash,omitempty"`

		// is running
		IsRunning bool `json:"isRunning,omitempty"`

		// path
		Path string `json:"path,omitempty"`

		// signer fingerprints
		SignerFingerprints []string `json:"signerFingerprints"`
	}
	buf := bytes.NewBuffer(raw)
	dec := json.NewDecoder(buf)
	dec.UseNumber()

	if err := dec.Decode(&data); err != nil {
		return err
	}

	var base struct {
		/* Just the base type fields. Used for unmashalling polymorphic types.*/

		ID *string `json:"id"`

		TypeID PostureCheckType `json:"typeId"`
	}
	buf = bytes.NewBuffer(raw)
	dec = json.NewDecoder(buf)
	dec.UseNumber()

	if err := dec.Decode(&base); err != nil {
		return err
	}

	var result PostureResponseProcessCreate

	result.idField = base.ID

	if base.TypeID != result.TypeID() {
		/* Not the type we're looking for. */
		return errors.New(422, "invalid typeId value: %q", base.TypeID)
	}

	result.Hash = data.Hash
	result.IsRunning = data.IsRunning
	result.Path = data.Path
	result.SignerFingerprints = data.SignerFingerprints

	*m = result

	return nil
}

// MarshalJSON marshals this object with a polymorphic type to a JSON structure
func (m PostureResponseProcessCreate) MarshalJSON() ([]byte, error) {
	var b1, b2, b3 []byte
	var err error
	b1, err = json.Marshal(struct {

		// hash
		Hash string `json:"hash,omitempty"`

		// is running
		IsRunning bool `json:"isRunning,omitempty"`

		// path
		Path string `json:"path,omitempty"`

		// signer fingerprints
		SignerFingerprints []string `json:"signerFingerprints"`
	}{

		Hash: m.Hash,

		IsRunning: m.IsRunning,

		Path: m.Path,

		SignerFingerprints: m.SignerFingerprints,
	})
	if err != nil {
		return nil, err
	}
	b2, err = json.Marshal(struct {
		ID *string `json:"id"`

		TypeID PostureCheckType `json:"typeId"`
	}{

		ID: m.ID(),

		TypeID: m.TypeID(),
	})
	if err != nil {
		return nil, err
	}

	return swag.ConcatJSON(b1, b2, b3), nil
}

// Validate validates this posture response process create
func (m *PostureResponseProcessCreate) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *PostureResponseProcessCreate) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID()); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this posture response process create based on the context it is used
func (m *PostureResponseProcessCreate) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *PostureResponseProcessCreate) contextValidateTypeID(ctx context.Context, formats strfmt.Registry) error {

	if err := m.TypeID().ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("typeId")
		} else if ce, ok := err.(*errors.CompositeError); ok {
			return ce.ValidateName("typeId")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *PostureResponseProcessCreate) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *PostureResponseProcessCreate) UnmarshalBinary(b []byte) error {
	var res PostureResponseProcessCreate
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
