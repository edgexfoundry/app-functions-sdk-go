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

package identity

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetIdentityEnrollmentsParams creates a new GetIdentityEnrollmentsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetIdentityEnrollmentsParams() *GetIdentityEnrollmentsParams {
	return &GetIdentityEnrollmentsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetIdentityEnrollmentsParamsWithTimeout creates a new GetIdentityEnrollmentsParams object
// with the ability to set a timeout on a request.
func NewGetIdentityEnrollmentsParamsWithTimeout(timeout time.Duration) *GetIdentityEnrollmentsParams {
	return &GetIdentityEnrollmentsParams{
		timeout: timeout,
	}
}

// NewGetIdentityEnrollmentsParamsWithContext creates a new GetIdentityEnrollmentsParams object
// with the ability to set a context for a request.
func NewGetIdentityEnrollmentsParamsWithContext(ctx context.Context) *GetIdentityEnrollmentsParams {
	return &GetIdentityEnrollmentsParams{
		Context: ctx,
	}
}

// NewGetIdentityEnrollmentsParamsWithHTTPClient creates a new GetIdentityEnrollmentsParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetIdentityEnrollmentsParamsWithHTTPClient(client *http.Client) *GetIdentityEnrollmentsParams {
	return &GetIdentityEnrollmentsParams{
		HTTPClient: client,
	}
}

/* GetIdentityEnrollmentsParams contains all the parameters to send to the API endpoint
   for the get identity enrollments operation.

   Typically these are written to a http.Request.
*/
type GetIdentityEnrollmentsParams struct {

	/* ID.

	   The id of the requested resource
	*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get identity enrollments params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetIdentityEnrollmentsParams) WithDefaults() *GetIdentityEnrollmentsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get identity enrollments params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetIdentityEnrollmentsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) WithTimeout(timeout time.Duration) *GetIdentityEnrollmentsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) WithContext(ctx context.Context) *GetIdentityEnrollmentsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) WithHTTPClient(client *http.Client) *GetIdentityEnrollmentsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) WithID(id string) *GetIdentityEnrollmentsParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get identity enrollments params
func (o *GetIdentityEnrollmentsParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *GetIdentityEnrollmentsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
