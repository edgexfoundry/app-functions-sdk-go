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

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/openziti/edge-api/rest_model"
)

// ListServiceServicePoliciesReader is a Reader for the ListServiceServicePolicies structure.
type ListServiceServicePoliciesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListServiceServicePoliciesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListServiceServicePoliciesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewListServiceServicePoliciesBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewListServiceServicePoliciesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 429:
		result := NewListServiceServicePoliciesTooManyRequests()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 503:
		result := NewListServiceServicePoliciesServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListServiceServicePoliciesOK creates a ListServiceServicePoliciesOK with default headers values
func NewListServiceServicePoliciesOK() *ListServiceServicePoliciesOK {
	return &ListServiceServicePoliciesOK{}
}

/* ListServiceServicePoliciesOK describes a response with status code 200, with default header values.

A list of service policies
*/
type ListServiceServicePoliciesOK struct {
	Payload *rest_model.ListServicePoliciesEnvelope
}

func (o *ListServiceServicePoliciesOK) Error() string {
	return fmt.Sprintf("[GET /services/{id}/service-policies][%d] listServiceServicePoliciesOK  %+v", 200, o.Payload)
}
func (o *ListServiceServicePoliciesOK) GetPayload() *rest_model.ListServicePoliciesEnvelope {
	return o.Payload
}

func (o *ListServiceServicePoliciesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.ListServicePoliciesEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceServicePoliciesBadRequest creates a ListServiceServicePoliciesBadRequest with default headers values
func NewListServiceServicePoliciesBadRequest() *ListServiceServicePoliciesBadRequest {
	return &ListServiceServicePoliciesBadRequest{}
}

/* ListServiceServicePoliciesBadRequest describes a response with status code 400, with default header values.

The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information
*/
type ListServiceServicePoliciesBadRequest struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceServicePoliciesBadRequest) Error() string {
	return fmt.Sprintf("[GET /services/{id}/service-policies][%d] listServiceServicePoliciesBadRequest  %+v", 400, o.Payload)
}
func (o *ListServiceServicePoliciesBadRequest) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *ListServiceServicePoliciesBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceServicePoliciesUnauthorized creates a ListServiceServicePoliciesUnauthorized with default headers values
func NewListServiceServicePoliciesUnauthorized() *ListServiceServicePoliciesUnauthorized {
	return &ListServiceServicePoliciesUnauthorized{}
}

/* ListServiceServicePoliciesUnauthorized describes a response with status code 401, with default header values.

The supplied session does not have the correct access rights to request this resource
*/
type ListServiceServicePoliciesUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceServicePoliciesUnauthorized) Error() string {
	return fmt.Sprintf("[GET /services/{id}/service-policies][%d] listServiceServicePoliciesUnauthorized  %+v", 401, o.Payload)
}
func (o *ListServiceServicePoliciesUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *ListServiceServicePoliciesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceServicePoliciesTooManyRequests creates a ListServiceServicePoliciesTooManyRequests with default headers values
func NewListServiceServicePoliciesTooManyRequests() *ListServiceServicePoliciesTooManyRequests {
	return &ListServiceServicePoliciesTooManyRequests{}
}

/* ListServiceServicePoliciesTooManyRequests describes a response with status code 429, with default header values.

The resource requested is rate limited and the rate limit has been exceeded
*/
type ListServiceServicePoliciesTooManyRequests struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceServicePoliciesTooManyRequests) Error() string {
	return fmt.Sprintf("[GET /services/{id}/service-policies][%d] listServiceServicePoliciesTooManyRequests  %+v", 429, o.Payload)
}
func (o *ListServiceServicePoliciesTooManyRequests) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *ListServiceServicePoliciesTooManyRequests) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListServiceServicePoliciesServiceUnavailable creates a ListServiceServicePoliciesServiceUnavailable with default headers values
func NewListServiceServicePoliciesServiceUnavailable() *ListServiceServicePoliciesServiceUnavailable {
	return &ListServiceServicePoliciesServiceUnavailable{}
}

/* ListServiceServicePoliciesServiceUnavailable describes a response with status code 503, with default header values.

The request could not be completed due to the server being busy or in a temporarily bad state
*/
type ListServiceServicePoliciesServiceUnavailable struct {
	Payload *rest_model.APIErrorEnvelope
}

func (o *ListServiceServicePoliciesServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /services/{id}/service-policies][%d] listServiceServicePoliciesServiceUnavailable  %+v", 503, o.Payload)
}
func (o *ListServiceServicePoliciesServiceUnavailable) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *ListServiceServicePoliciesServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
