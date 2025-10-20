package rest

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Meta struct {
	TotalSize     int32  `json:"total_size,omitempty"`
	NextPageToken string `json:"next_page_token"`
}

type Response struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

type Data struct {
	Data interface{} `json:"data"`
}

// Respond writes response to http.ResponseWriter.
// It to the user to build the response struct
func Respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("unable to encode response data with error: ", err)
	}
}

// RespondWithResponse writes the project standard response {data, meta} to http.ResponseWriter.
// This should be used unless there is a unique situation in which the standard can't be used
func RespondWithResponse(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	res := Response{
		Data: data,
		Meta: meta,
	}

	Respond(w, status, res)
}

// RespondStatusCreated removes the boiler plate of creating the response struct.
func RespondStatusCreated(w http.ResponseWriter, id string) {
	idJSON := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	RespondWithResponse(w, http.StatusCreated, idJSON, nil)
}

// RespondNoContent acknowleges success but has nothing to return
func RespondNoContent(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// jsonError handles the specific response written to http.ResponseWriter
func jsonError(w http.ResponseWriter, status int, code, msg string, fields []FieldError) {
	w.WriteHeader(status)
	w.Header().Add("content-Type", "application/json")

	err := ErrorResponse{
		Error: &ErrorPayload{
			Code:    code,
			Message: msg,
			Fields:  fields,
		},
	}
	// log here
	encErr := json.NewEncoder(w).Encode(err)
	if encErr != nil {
		log.Println(err)
	}
}

// JSONHandleError
func RespondJSONError(w http.ResponseWriter, err error) {
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		apiErr = Err500Default
	}

	status, code, msg, fields := apiErr.GetData()
	jsonError(w, status, code, msg, fields)
}

// RespondProto writes a proto struct response to http.ResponseWriter.
// using json struct tags within proto has been deprecated and only populates to ensure backwards compatibility
// the correct way is to handle the unmarshalling ourselves which allows us to custom options without a proto tag plugin
// this is used for list endpoints as we currently build the meta as part of the response proto
func RespondProto(w http.ResponseWriter, status int, data proto.Message) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	// set the custom marshalling config
	protomarshal := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	jsonpb, err := protomarshal.Marshal(data)
	if err != nil {
		log.Println("unable to marshal response proto data with error: ", err)
	}

	_, err = w.Write(jsonpb)
	if err != nil {
		log.Println("unable to encode response proto data with error: ", err)
	}
}

// RespondWithProtoResponse writes the project standard response {data, meta} to http.ResponseWriter.
// This should be used unless there is a unique situation in which the standard proto type are not being used
func RespondWithProtoResponse(w http.ResponseWriter, status int, data proto.Message, meta interface{}) {
	w.Header().Add("Content-Type", "application/json")

	// Marshal the proto data
	protomarshal := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	jsonpb, err := protomarshal.Marshal(data)
	if err != nil {
		log.Println("unable to marshal proto data with error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Unmarshal proto JSON to a generic interface for manipulation
	var dataMap interface{}
	err = json.Unmarshal(jsonpb, &dataMap)
	if err != nil {
		log.Println("unable to unmarshal proto json:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Construct the standard response {data, meta}
	response := map[string]interface{}{
		"data": dataMap,
		"meta": meta,
	}

	// Marshal the response structure to JSON
	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Println("unable to marshal response with error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write the status code and response
	w.WriteHeader(status)
	_, err = w.Write(responseJson)
	if err != nil {
		log.Println("unable to write response:", err)
	}
}
